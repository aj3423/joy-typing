package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aj3423/joy_typing/joycon"
	"github.com/aj3423/joy_typing/mode"
	"github.com/arturoeanton/go-notify"
	"github.com/pelletier/go-toml/v2"
	log "github.com/sirupsen/logrus"
)

const ConfigFile = "config.toml"

type Config struct {
	LogLevel             log.Level `comment:"panic,fatal,error,warn,info,debug,trace"`
	SpinNeutralThreshold float64   `comment:"Stick is considered as 'neutral' if the spinning ratio is below this percentage (range: 0~1.0)"`
	SpinEdgeThreshold    float64   `comment:"Stick Up/Down/Left/Right events are triggered when the spinning ratio exceeds this value (range: 0~1.0)"`
	// use these 3 simple structs instead of embed other struct,
	// because that would result in a complex layout in config file.
	ModeList    []mode.ModeConfig   `toml:"Mode,multiline" comment:"rules for all modes"`
	PhraseList  map[string][]string `toml:"PhraseList,multiline" comment:"This section is used to narrow down the word dictionary of a speech mode,\n used as parameter '-phrase' of 'speech mode', can appear multiple times,\n for example: '[speech] -id MyGolangMode -phrase common application java lua'"`
	WordMapping map[string][]string `toml:"WordMapping,multiline" comment:"Can't figure out how to display the items below in multiline, just format it with some online formatter and copy back:-)"`
}

func loadConfig() (e error) {
	s, e := ioutil.ReadFile(ConfigFile)
	if errors.Is(e, os.ErrNotExist) { // generate at first launch
		// generate a .json and use default settings
		if e = saveConfig(); e != nil {
			return e
		}
		s, e = ioutil.ReadFile(ConfigFile)
	}

	if e != nil {
		return e
	}
	// have to use a new `Config` instance for `toml.Unmarshal()`,
	// otherwise it panics:
	var cfg = &Config{}
	if e = toml.Unmarshal(s, cfg); e != nil {
		return e
	}
	// save to cfg, it's used when `saveConfig()`
	currCfg = cfg

	// apply config to all packages
	mode.ModeList = currCfg.ModeList
	mode.PhraseList = currCfg.PhraseList
	mode.WordMapping = currCfg.WordMapping
	joycon.SpinNeutralThreshold = currCfg.SpinNeutralThreshold
	joycon.SpinEdgeThreshhold = currCfg.SpinEdgeThreshold
	log.SetLevel(currCfg.LogLevel)

	modes, modeSitches, e := mode.Parse()
	if e != nil {
		return fmt.Errorf("failed to parse mode: %s", e.Error())
	}
	if e := mode.Manager.SetModes(modes, modeSitches); e != nil {
		return fmt.Errorf("failed to parse mode: %s", e.Error())
	}
	return nil
}

// reload config when the file is modified
func watchConfig() chan struct{} {
	stop := make(chan struct{})
	go func() {
		fx := func(_ *notify.ObserverNotify, _ *notify.Event) {
			e := loadConfig()
			if e == nil {
				log.Info("config reloaded successfully")
			} else {
				log.Errorf("Failed to reload config: %s", e.Error())
			}
		}

		notify.NewObserverNotify("./", ConfigFile).
			FxWrite(fx).
			Run()

		<-stop
	}()
	return stop
}

func saveConfig() error {
	buf := bytes.Buffer{}
	enc := toml.NewEncoder(&buf)
	enc.SetIndentTables(true)
	enc.Encode(currCfg)

	e := ioutil.WriteFile(ConfigFile, buf.Bytes(), 0644)
	if e == nil {
		log.Infof("'%s' generated, it's a demo for how to use the Right Joy-Con", ConfigFile)
	}
	return e
}
