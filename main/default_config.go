package main

import (
	"github.com/aj3423/joy-typing/joycon"
	"github.com/aj3423/joy-typing/mode"
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
				`[trigger] button -id Y -> [speak] -text space`,

				`[trigger] button -id R-SR -> [hotkey] -keys s ctrl`,
				`[trigger] button -id Home -> [repeat]`,

				// stick
				`[trigger] stick -side Right -dir Up    -> [hotkey] -keys up`,
				`[trigger] stick -side Right -dir Down  -> [hotkey] -keys down`,
				`[trigger] stick -side Right -dir Left  -> [hotkey] -keys left`,
				`[trigger] stick -side Right -dir Right -> [hotkey] -keys right`,

				`[trigger] speech -> [speech] -map programming vim application go`,

				// to verfy if the gyro has been turned off after leaving mouse mode,
				// if the packet lost and gyro fails to stop, it can trigger mouse move in this mode
				`[trigger] gyro -> [cursor] -speed 0.03`,
			},
		},
		{
			Mode: `[speech] -id WordMode -flushonexit -phrase programming vim go application test`,
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
				`[trigger] speech -> [speech] -map programming vim application go`,
			},
		},
		{
			Mode: `[speech] -id SentenceMode`,
			Rules: []string{
				`[switch] button -id X   -> [upper]`,

				// stick
				`[trigger] stick -side Right -dir Up    -> [hotkey] -keys up`,
				`[trigger] stick -side Right -dir Down  -> [hotkey] -keys down`,
				`[trigger] stick -side Right -dir Left  -> [hotkey] -keys left`,
				`[trigger] stick -side Right -dir Right -> [hotkey] -keys right`,

				`[trigger] speech        -> [speech] --nospace=false -map programming application go`,
			},
		},
		{
			Mode: `[gyro] -id MouseMode`,
			Rules: []string{
				`[trigger] gyro -> [cursor] -speed 0.03`,
				`[switch]  button -id R   -> [mouse_toggle]`,
				`[trigger] button -id ZR  -> [click]`,
				`[trigger] button -id X   -> [click] -button right`,
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
		`programming`: {
			`a`, `alpha`, `b`, `bat`, `c`, `d`, `e`, `emma`, `f`, `g`, `h`, `i`, `j`, `k`, `l`, `lot`, `m`, `maiden`,
			`n`, `near`, `o`, `p`, `q`, `r`, `s`, `t`, `u`, `v`, `vest`, `w`, `x`, `y`, `z`, `zed`,

			`zero`, `one`, `two`, `three`, `four`, `five`, `six`, `seven`, `eight`, `nine`, `ten`,
			`eleven`, `twelve`, `thirteen`, `fourteen`, `fifteen`, `sixteen`, `seventeen`, `eighteen`, `nineteen`,
			`twenty`, `thirty`, `forty`, `fifty`, `sixty`, `seventy`, `eighty`, `ninety`,
			`hundred`, `thousand`,

			`second`, `millisecond`, `minute`, `hour`, `day`, `weak`, `month`, `year`,

			`upper`, `camel`, `title`, `snake`,

			`slash`, `underline`, `underscore`, `dash`, `negative`, `plus`, `or`,
			`not`, `equal`, `dot`, `point`, `comma`, `semi`, `colon`, `semicolon`, `double`, `single`, `loop`, `condition`,
			`single`, `quote`, `quotes`, `tick`, `back`, `double`, `round`, `square`, `curly`, `angle`, `bracket`, `brackets`, `close`, `pair`,
			`return`, `enter`, `space`, `bar`, `left`, `right`, `up`, `down`, `caret`, `at`, `tidal`, `question`,
			`minus`, `plus`, `multiply`, `divide`, `star`, `comment`, `dollar`, `mod`, `escape`, `less`, `greater`, `than`,
			`repeat`, `append`, `pre`, `post`, `begin`, `ending`, `check`,
			// conflicts:
			// `arrow`<>`l`
			// `while`<>`y`
			// `case`<>`k`
		},
		`application`: {
			`hotkey`, `launch`, `shell`, `bash`, `control`, `shift`, `alt`, `tab`, `meta`, `move`, `resize`, `maximize`, `minimize`, `window`,
			`run`, `command`, `telegram`, `chrome`, `brave`, `terminal`, `aptitude`, `nala`, `vim`, `explorer`, `update`, `upgrade`, `vim`, `shut`,
			`close`, `workspace`, `switch`, `git`, `clone`, `status`, `file`, `manager`, `job`, `note`,
		},
		`test`: {
			`elephant`, `cobra`, `delay`, `hello`, `world`,
		},
		`go`: {
			`struct`, `var`, `type`, `import`, `package`, `switch`, `function`, `go`, `make`, `new`, `const`, `unsigned`, `int`, `float`, `byte`, `array`,
			`spring`, `print`, `line`, `format`, `assert`, `select`, `main`, `truth`, `false`, `break`,
		},
		`vim`: {
			`mode`, `insert`, `undo`, `redo`, `search`, `mark`, `next`, `previous`, `last`, `position`, `modify`, `zoom`, `page`, `middle`, `top`, `bottom`, `down`, `buffer`,
		},
	},
	WordMapping: map[string][]string{
		`vim`: {
			`s_h -> [hotkey] h shift`, `c_o -> [hotkey] o ctrl`, `c_i -> [hotkey] i ctrl`, `insert -> i`, `select line -> escape vil`,
			`select 9 -> escape vil`, `line begin -> c_o I`, `9 begin -> c_o I`, `line ending -> c_o A`, `9 ending -> c_o A`, `undo -> escape u`,
			`redo -> escape ";redo" enter`, `new line -> c_o o`, `new 9 -> c_o o`, `new line upper -> c_o O`, `new 9 upper -> c_o O`, `search -> "/"`,
			`mark -> m`, `next -> n`, `previous -> N`, `last position -> c_o c_o`, `next position -> c_o c_i`, `last modify -> c_o "'" "."`,
			`middle -> M`, `top -> H`, `bottom -> L`, `zoom -> zz `, `page down -> [hotkey] d ctrl`, `page up -> [hotkey] e ctrl`,
			`buffer right -> [hotkey] l ctrl`, `buffer left -> [hotkey] h ctrl`, `buffer up -> [hotkey] k ctrl`, `buffer down -> [hotkey] j ctrl`,
		},
		`programming`: {
			`repeat -> [repeat]`,
			`the -> ""`, `alpha -> "a"`, `bat -> "b"`, `emma -> "e"`, `lot -> "l"`, `maiden -> "m"`, `near -> "n"`, `vest -> "v"`, `zed -> z`, `escape -> [hotkey] esc`, `space -> " "`,
			`bar -> " "`, `dot -> "."`, `equal -> =`, `dash -> "-"`, `minus -> "-"`, `negative -> "-"`, "tick -> \"`\"", `plus -> "+"`, `at -> "@"`, `multiply -> "*"`,
			`star -> "*"`, `divide -> "/"`, `dollar -> "$"`, `underscore -> _`, `underline -> _`, `enter -> [hotkey] enter`, `return -> [hotkey] enter`, `round -> "("`,
			`round close -> ")"`, `round pair-> "()"`, `square -> "["`, `square close -> "]"`, `square pair -> "[]"`, `curly -> "{"`, `curly close -> "}"`, `curly pair -> "{}"`,
			`angle -> "<"`, `angle close -> ">"`, `angle pair -> "<>"`, `less than -> "<"`, `greater than -> ">"`, `less equal -> " <= "`, `greater equal -> " >= "`,
			`single quote -> "'"`, `quote -> \"`, `semi -> ";"`, `colon -> ":"`, `comma -> ","`, `f check -> "if "`, `y loop -> "while "`,

			`camel -> [camel]`, `title -> [title]`, `snake -> [snake]`, `upper -> [upper]`,
			`elephant -> [camel] [title]`, // ThisIsElephant
			`cobra -> [snake] [title]`,    // I_am_cobra-_-
		},
		`application`: {
			`file manager -> [shell] dw`,
			`run terminal -> [hotkey] t control alt`,
			`run job -> [shell] v "/e/job/job.md"`,
			`run note -> [shell] v "/e/job/note.md"`,
			`run brave -> [shell] "brave-browser" "--no-sandbox"`,
			`run command -> [hotkey] esc alt`,
			`maximize window -> [hotkey] e alt`,
			`minimize window -> [hotkey] e alt`,
			`shut -> [hotkey] w alt`,
			`resize window -> [hotkey] r alt`,
			`move window -> [hotkey] e alt`,
			`control c -> [hotkey] c ctrl`,
			`control v -> [hotkey] v ctrl`,
			`control d -> [hotkey] d ctrl`,
			`control p -> [hotkey] p ctrl`,
			`control f -> [hotkey] f ctrl`,
			`control zed -> [hotkey] z ctrl`,
			`new tab -> [hotkey] t ctrl`,
			`last page -> [hotkey] left alt`,
			`next page -> [hotkey] right alt`,
			`close tab -> [hotkey] w ctrl`,
			`move workspace 2 -> [hotkey] 2 alt ctrl`,
			`move workspace 1 -> [hotkey] 1 alt ctrl`,
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
