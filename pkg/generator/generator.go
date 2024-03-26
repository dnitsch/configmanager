package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/dnitsch/configmanager/internal/config"
	"github.com/dnitsch/configmanager/internal/store"
	"github.com/dnitsch/configmanager/pkg/log"
	"github.com/spyzhov/ajson"
)

type muRawMap struct {
	mu       sync.RWMutex
	tokenMap ParsedMap
}

// GenVars is the main struct holding the
// strategy patterns iface
// any initialised config if overridded with withers
// as well as the final outString and the initial rawMap
// which wil be passed in a loop into a goroutine to perform the
// relevant strategy network calls to the config store implementations
type GenVars struct {
	ctx    context.Context
	config config.GenVarsConfig
	// rawMap is the internal object that holds the values
	// of original token => retrieved value - decrypted in plain text
	// with a mutex RW locker
	rawMap muRawMap //ParsedMap
}

// ParsedMap is the internal working object definition and
// the return type if results are not flushed to file
type ParsedMap map[string]any

// NewGenerator returns a new instance of Generator
// with a default strategy pattern wil be overwritten
// during the first run of a found tokens map
func NewGenerator() *GenVars {
	// defaultStrategy := NewDefatultStrategy()
	return newGenVars()
}

func newGenVars() *GenVars {
	m := make(ParsedMap)
	return &GenVars{
		rawMap: muRawMap{tokenMap: m},
		ctx:    context.TODO(),
		// return using default config
		config: *config.NewConfig(),
	}
}

// WithConfig uses custom config
func (c *GenVars) WithConfig(cfg *config.GenVarsConfig) *GenVars {
	// backwards compatibility
	if cfg != nil {
		c.config = *cfg
	}
	return c
}

// WithContext uses caller passed context
func (c *GenVars) WithContext(ctx context.Context) *GenVars {
	c.ctx = ctx
	return c
}

// Config gets Config on the GenVars
func (c *GenVars) Config() *config.GenVarsConfig {
	return &c.config
}

func (c *GenVars) RawMap() ParsedMap {
	c.rawMap.mu.RLock()
	defer c.rawMap.mu.RUnlock()
	// make a copy of the map
	m := make(ParsedMap)
	for k, v := range c.rawMap.tokenMap {
		m[k] = v
	}
	return m
}

func (c *GenVars) AddRawMap(key *config.ParsedTokenConfig, val string) {
	c.rawMap.mu.Lock()
	defer c.rawMap.mu.Unlock()
	// NOTE: still use the metadata in the key
	// there could be different versions / labels for the same token and hence different values
	// However the JSONpath look up
	c.rawMap.tokenMap[key.String()] = c.keySeparatorLookup(key.StripMetadata(), val)
}

type rawTokenMap map[string]*config.ParsedTokenConfig

// Generate generates a k/v map of the tokens with their corresponding secret/paramstore values
// the standard pattern of a token should follow a path like string
func (c *GenVars) Generate(tokens []string) (ParsedMap, error) {

	parsedTokenMap := make(rawTokenMap)
	for _, token := range tokens {
		// TODO: normalize tokens here potentially
		// merge any tokens that only differ in keys lookup inside the object
		parsedToken := config.NewParsedTokenConfig(token, c.config)
		if parsedToken != nil {
			parsedTokenMap[token] = parsedToken
		}
	}
	rs := NewRetrieveStrategy(store.NewDefatultStrategy(), c.config)
	// pass in default initialised retrieveStrategy
	// input should be
	if err := c.generate(parsedTokenMap, rs); err != nil {
		return nil, err
	}
	return c.RawMap(), nil
}

type retrieveIface interface {
	RetrieveByToken(ctx context.Context, impl store.Strategy, in *config.ParsedTokenConfig) *TokenResponse
	SelectImplementation(ctx context.Context, in *config.ParsedTokenConfig) (store.Strategy, error)
}

// generate checks if any tokens found
// initiates groutines with fixed size channel map
// to capture responses and errors
// generates ParsedMap which includes
func (c *GenVars) generate(rawMap rawTokenMap, rs retrieveIface) error {
	if len(rawMap) < 1 {
		log.Debug("no replaceable tokens found in input strings")
		return nil
	}

	var errors []error
	// build an exact size channel
	var wg sync.WaitGroup
	initChanLen := len(rawMap)
	outCh := make(chan *TokenResponse, initChanLen)

	wg.Add(initChanLen)
	// TODO: initialise the singleton serviceContainer
	// pass into each goroutine
	for _, parsedToken := range rawMap {
		// take value from config allocation on a per iteration basis
		go func(tkn *config.ParsedTokenConfig) {
			defer wg.Done()
			strategy, err := rs.SelectImplementation(c.ctx, tkn)
			if err != nil {
				outCh <- &TokenResponse{Err: err}
				return
			}
			outCh <- rs.RetrieveByToken(c.ctx, strategy, tkn)
		}(parsedToken)
	}

	go func() {
		wg.Wait()
		close(outCh)
	}()

	for cro := range outCh {
		cr := cro
		log.Debugf("cro: %+v", cr)
		if cr.Err != nil {
			log.Debugf("cr.err %v, for token: %s", cr.Err, cr.key)
			errors = append(errors, cr.Err)
			// Skip adding not found key to the RawMap
			continue
		}
		c.AddRawMap(cr.key, cr.value)
	}

	if len(errors) > 0 {
		// crude ...
		log.Debugf("found: %d errors", len(errors))
		// return outMap, fmt.Errorf("%v", errors)
	}
	return nil
}

// IsParsed will try to parse the return found string into
// map[string]string
// If found it will convert that to a map with all keys uppercased
// and any characters
func IsParsed(res any, trm *ParsedMap) bool {
	str := fmt.Sprint(res)
	if err := json.Unmarshal([]byte(str), &trm); err != nil {
		log.Info("unable to parse into a k/v map returning a string instead")
		return false
	}
	log.Info("parsed into a k/v map")
	return true
}

// keySeparatorLookup checks if the key contains
// keySeparator character
// If it does contain one then it tries to parse
func (c *GenVars) keySeparatorLookup(key, val string) string {
	// key has separator
	kl := strings.Split(key, c.config.KeySeparator())
	log.Debugf("key list: %v", kl)
	if len(kl) < 2 {
		log.Infof("no keyseparator found")
		return val
	}

	keys, err := ajson.JSONPath([]byte(val), fmt.Sprintf("$..%s", kl[1]))
	if err != nil {
		log.Debugf("unable to parse as json object %v", err.Error())
		return val
	}

	if len(keys) == 1 {
		v := keys[0]
		if v.Type() == ajson.String {
			str, err := strconv.Unquote(fmt.Sprintf("%v", v))
			if err != nil {
				log.Debugf("unable to unquote value: %v returning as is", v)
				return fmt.Sprintf("%v", v)
			}
			return str
		}

		return fmt.Sprintf("%v", v)
	}

	log.Infof("no value found in json using path expression")
	return ""
}
