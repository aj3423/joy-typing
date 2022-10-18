package main

import (
	"github.com/aj3423/joy_typing/joycon"
	"github.com/aj3423/joy_typing/mode"
	log "github.com/sirupsen/logrus"
)

// global variable
var currCfg = &Config{
	LogLevel:             log.InfoLevel,
	SpinNeutralThreshold: joycon.SpinNeutralThreshold,
	SpinEdgeThreshold:    joycon.SpinEdgeThreshhold,
	ModeList: []mode.ModeConfig{
		{
			Mode: `[idle] -id id1`,
			Rules: []string{

				// modes
				`[switch]  button -id ZR   -> [mode] -id WordMode`,
				`[switch]  button -id R    -> [mode] -id SentenceMode`,
				`[switch]  button -id "+"  -> [mode] -id MouseMode`,
				`[switch]  button -id R-SL -> [mode] -id some_test_mode`,

				// buttons
				`[trigger] button -id X -> [hotkey] -keys i`,
				`[trigger] button -id B -> [hotkey] -keys esc`,
				`[trigger] button -id A -> [hotkey] -keys enter`,

				`[trigger] button -id R-SR -> [hotkey] -keys s ctrl`,
				`[trigger] button -id Home -> [repeat]`,

				// stick
				`[trigger] stick -side Right -dir Up    -> [hotkey] -keys up`,
				`[trigger] stick -side Right -dir Down  -> [hotkey] -keys down`,
				`[trigger] stick -side Right -dir Left  -> [hotkey] -keys left`,
				`[trigger] stick -side Right -dir Right -> [hotkey] -keys right`,

				`[trigger] speech -> [speech] -map common programming vim application go`,

				// to verfy if the gyro has been turned off after leaving mouse mode,
				// if the packet lost and gyro fails to stop, it can trigger mouse move in this mode
				`[trigger] gyro -> [cursor] -speed 0.03`,
			},
		},
		{
			Mode: `[speech] -id WordMode -flushonexit -phrase common vim go command application test`,
			Rules: []string{
				// buttons
				`[trigger] button -id X -> [speak] -text "c_o de"`,
				`[trigger] button -id B -> [speak] -text "c_o db"`,
				`[trigger] button -id Y -> [hotkey] -keys backspace`,
				`[trigger] button -id A -> [hotkey] -keys delete`,
				`[trigger] button -id Home -> [hotkey] -keys S`,

				// stick
				`[trigger] stick -side Right -dir Left  -> [speak] -text "c_o b"`,
				`[trigger] stick -side Right -dir Right -> [speak] -text "c_o e"`,
				// `[trigger] stick -side Right -dir Up    -> [speak] -text "c_o e"`,
				`[trigger] stick -side Right -dir Down  -> [speak] -text "c_o w"`,

				`[trigger] button -id Home  -> [notify] -text text -title title`,
				`[trigger] button -id B     -> [flush]`,
				`[trigger] speech -> [speech] -map common programming vim application go`,
			},
		},
		{
			Mode: `[speech] -id SentenceMode`,
			Rules: []string{
				`[switch] button -id X   -> [upper]`,
				`[trigger] speech        -> [speech] -map common programming application go`,
			},
		},
		{
			Mode: `[gyro] -id MouseMode`,
			Rules: []string{
				`[trigger] gyro -> [cursor] -speed 0.03`,
				`[switch]  button -id R    -> [mouse_toggle]`,
				`[trigger] button -id ZR   -> [click]`,

				`[trigger] button -id A   -> [click] -button right`,
				`[trigger] button -id B   -> [click] -double`,
				`[switch]  button -id "+" -> [boost] -multiplier 0.5`,
			},
		},
		{
			Mode: `[speech] -id some_test_mode`,
			Rules: []string{
				`[trigger] speech -> [speech] --nospace=false --number=false`,
				`[trigger] button -id R-SR  -> [notify] -text "just a test" -title "a title"`,
			},
		},
	},
	PhraseList: map[string][]string{
		`common`: {
			`a`, `alpha`, `b`, `bat`, `c`, `d`, `e`, `each`, `f`, `g`, `h`, `i`, `j`, `k`, `l`, `m`, `maiden`,
			`n`, `near`, `o`, `p`, `q`, `r`, `s`, `t`, `u`, `v`, `vest`, `w`, `x`, `y`, `z`, `zed`,

			`zero`, `one`, `two`, `three`, `four`, `five`, `six`, `seven`, `eight`, `nine`, `ten`,
			`eleven`, `twelve`, `thirteen`, `fourteen`, `fifteen`, `sixteen`, `seventeen`, `eighteen`, `nineteen`,
			`twenty`, `thirty`, `forty`, `fifty`, `sixty`, `seventy`, `eighty`, `ninety`,
			`hundred`, `thousand`,

			`second`, `millisecond`, `minute`, `hour`, `day`, `weak`, `month`, `year`,

			`upper`, `camel`, `title`, `snake`,

			`slash`, `underline`, `underscore`, `dash`, `negative`, `plus`, `or`,
			`not`, `equal`, `dot`, `point`, `comma`, `semi`, `colon`, `semicolon`, `double`, `single`, `loop`, `if`,
			`single`, `quote`, `quotes`, `tick`, `back`, `double`, `round`, `square`, `curly`, `angle`, `bracket`, `brackets`, `close`, `pair`,
			`return`, `enter`, `space`, `left`, `right`, `up`, `down`, `caret`, `at`, `tidal`, `question`,
			`minus`, `plus`, `multiply`, `divide`, `star`, `comment`, `dollar`, `mod`, `escape`, `less`, `greater`, `than`,
			`repeat`, `append`, `pre`, `post`, `begin`, `ending`,
			// conflicts:
			// `arrow`<>`l`
			// `while`<>`y`
			// `case`<>`k`
		},
		`command`: {
			`hotkey`, `launch`, `shell`, `bash`, `control`, `shift`, `alt`, `meta`,
		},
		`application`: {
			`telegram`, `chrome`, `brave`, `terminal`,
		},
		`test`: {
			`elephant`, `cobra`, `delay`, `hello`, `world`,
		},
		`go`: {
			`struct`, `var`, `type`, `import`, `package`, `switch`, `function`, `go`, `make`, `new`, `const`, `unsigned`, `int`, `float`, `byte`, `array`,
			`spring`, `print`, `line`, `format`, `assert`, `select`, `main`, `truth`, `false`,
		},
		`vim`: {
			`mode`, `insert`, `undo`, `redo`, `search`, `mark`, `next`, `previous`, `last`, `position`, `modify`, `zoom`, `middle`, `top`, `bottom`, `buffer`,
		},
	},
	WordMapping: map[string][]string{
		`vim`: {
			`s_h -> [hotkey] h shift`,
			`c_o -> [hotkey] o ctrl`,
			`c_i -> [hotkey] i ctrl`,
			`insert -> i`,
			`select line -> escape vil`,
			`select 9 -> escape vil`,
			`line begin -> c_o I`,
			`9 begin -> c_o I`,
			`line ending -> c_o A`,
			`9 ending -> c_o A`,
			`undo -> c_o u`,
			`redo -> c_o ";redo" enter`,
			`new line -> c_o o`,
			`new 9 -> c_o o`,
			`new line upper -> c_o O`,
			`new 9 upper -> c_o O`,
			`search -> "/"`,
			`mark -> m`,
			`next -> n`,
			`previous -> N`,
			`last position -> c_o c_o`,
			`next position -> c_o c_i`,
			`last modify -> c_o "'" "."`,
			`middle -> M`,
			`top -> H`,
			`bottom -> L`,
			`zoom -> zz `,
			`zoom next -> bottom zt`,
			`buffer right -> [hotkey] l ctrl`,
			`buffer left -> [hotkey] h ctrl`,
			`buffer up -> [hotkey] k ctrl`,
			`buffer down -> [hotkey] j ctrl`,
		},
		`common`: {
			`repeat -> [repeat]`,
		},
		`programming`: {
			`the -> ""`,
			`alpha -> "a"`,
			`bat -> "b"`,
			`each -> "e"`,
			`maiden -> "m"`,
			`near -> "n"`,
			`vest -> "v"`,
			`escape -> [hotkey] esc`,
			`space -> " "`,
			`dot -> "."`,
			`equal -> =`,
			`dash -> "-"`,
			`minus -> "-"`,
			`negative -> "-"`,
			"tick -> \"`\"",
			`plus -> "+"`,
			`at -> "@"`,
			`multiply -> "*"`,
			`star -> "*"`,
			`divide -> "/"`,
			`dollar -> "$"`,
			`underscore -> _`,
			`underline -> _`,
			`enter -> [hotkey] enter`,
			`return -> [hotkey] enter`,
			`round -> "("`,
			`round close -> ")"`,
			`round pair-> "()"`,
			`square -> "["`,
			`square close -> "]"`,
			`square pair -> "[]"`,
			`curly -> "{"`,
			`curly close -> "}"`,
			`curly pair -> "{}"`,
			`angle -> "<"`,
			`angle close -> ">"`,
			`angle pair -> "<>"`,
			`less than -> "<"`,
			`greater than -> ">"`,
			`less equal -> " <= "`,
			`greater equal -> " >= "`,
			`single quote -> "'"`,
			`quote -> \"`,
			`semi -> ";"`,
			`colon -> ":"`,
			`comma -> ","`,

			`camel -> [camel]`,
			`title -> [title]`,
			`snake -> [snake]`,
			`upper -> [upper]`,
			`elephant -> [camel] [title]`, // ThisIsElephant
			`cobra -> [snake] [title]`,    // I_am_cobra-_-
		},
		`application`: {
			`launch terminal -> [hotkey] t control alt`,
			`launch brave -> [shell] "brave-browser" "--no-sandbox" "www.test.com"`,
			`meta delay -> [delay] 1s`,
		},
		`go`: {
			`spring -> "fmt.Sprintf("`,
			`assert -> "if e != nil {"`,
			`4 loop -> "for "`,
			`colon equal -> " := "`,
			`function -> "func "`,
			`truth -> true`,
		},
	},
}
