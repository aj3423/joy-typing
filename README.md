![logo](https://user-images.githubusercontent.com/4710875/196342674-f568755f-d371-44db-9c56-663044d786bc.png)


## Joy Typing
**Use your computer with Voice and Joy-Con**

- Replace `Keyboard` -> `Voice Typing` + `Joy-Con Button`
- Replace `Mouse` -> `Motion Control` + `Joy-Con Stick`

**For those who wants to:**
 - Avoid maintaining the typing posture.
 - Play PC game with motion control.


## How it works
**1. The Mode concept**
 The idea comes from Vim, there could be different kinds of modes: *normal*, *speech*, *motion control*. The Joy-Con has a very limited number of buttons, but the same button can trigger different actions in different modes, hence it's greatly extended.

**2. Word Mapping**
 Words can be mapped to different actions like:
 - Replacement: `dash` -> `-`, `twenty twenty two` -> `2022`
 - Decoration: `snake hello world` -> `hello_world`, `camel hello world` -> `helloWorld`
 - Hotkey: `control alt delete` -> `task explorer launched`
 - Run shell script: `launch browser` -> `brave-browser --no-sandbox www.test.com`

**3. Limit the dictionary for better accuracy**
  [VOSK](https://alphacephei.com/vosk/ "VOSK") is used as the backend recognition engine, with the [phrase_list](https://github.com/alphacep/vosk-server/blob/master/websocket/asr_server.py#L44 "phrase_list") parameter, it's possible to use a small dictionary, for instance, when the dictionary is limited to alphabet`a~z`,  the`c`will never be recognized as`sea`or`see`.
  For programming, switch to a speech mode with limited dictionary for typing keywords, punctuation, numbers. And use another mode with unlimited dictionary for variables and comments, once found a conflict, solve it by adding a word mapping.

##  Install
**1. Connect Joy-Con to PC via Bluetooth**
- From the system bluetooth manager, click `scan new device`
- Hold the sync button for 2 seconds, until the lights start blinking.

![sync_button](https://user-images.githubusercontent.com/4710875/196342778-a4a9d074-eb49-4af2-8ce0-bb3add94bf28.png)

- Find -> Pair -> Connect it in the BT manager.

**2. Install Docker**
 -  Windows
Download and run the installer from https://docs.docker.com/desktop/install/windows-install/. If it prompt something like `download and install WSL 2 Linux kernel upgrade package`, just install it.

 - Linux
`apt install docker.io`

 - Mac
I have zero experience with Mac, also there is no prebuilt binary for Mac, but the steps should be similar to Windows/Linux, all the dependency packages claim to support Mac, check the next section to build it from souce.

**3. Run VOSK server**

`docker run -d -p 2701:2701 aj3423/vosk_lgraph:latest`

**4. Download the prebuilt binary from release page**


## Build from Source
1. install dependency `hidapi`
 - Windows: Download the release file from https://github.com/libusb/hidapi/releases, which contains header, lib and dll.
 - Linux: `apt install libhidapi-dev`

2. Install golang from https://go.dev/dl/
3. Clone this repo: `git clone github.com/aj3423/joy-typing`
4. Go to 'main' directory: `cd joy-typing/main`
5. `go build .`


## Troubleshooting
1. **Keeps disconnecting right after connected**

The Joy-Con can **ONLY** be paired to one device at a time, once you attach it back to the switch console for charging, it's auto re-paired to the console, you'll have to remove it in the system BT Manager and re-pair it again. I tried some hacky way like attach it to the console during the shutting down or powering up, to get it being charged but not re-paired, I succeeded only once by accident but can't remember how, ended up using some dedicated charging cable.

2. **No sound input or inaccurate recognition**

Diagnose with this tool: [vosk-sound-test](https://github.com/aj3423/vosk-sound-test "vosk-sound-test")
It's capable of playing back your voice and saving to a .wav file, so you can verify if the sound quality is expected.
Usage:
 1). `docker run -d -p 2701:2701 aj3423/vosk_lgraph:latest`
 2) Download the binary from the [release page](https://github.com/aj3423/vosk-sound-test/releases/tag/1.0 "release page")
 3) `./vosk-sound-test -host "127.0.0.1:2701"`
 4) Say something, press enter to playback, press enter again to save to a .wav file.

3. **Unexpected behavior of Joy-Con**

When press both side buttons(`SL` + `SR`), Joy-Con enters an interesting mode, maybe it's the mode when being attached. Check the lights, if the first light keeps on and other three are off, just re-connect it and make sure not press both side buttons. BTW, in this mode, the buttons are bound to some system operations, for example the Joy-Con-Right:

	- +: toggle on/off the events below
	- R: mouse right click
	- B: `Esc`
	- Home: system volumn down
	- Stick button: system volumn up
	- Stick spin: mouse move
	...


## Configuration
The file `config.toml` is generated at the first launch, it monitors file modification and applys new changes on the fly. The sections:

**1. Mode**

A mode id must be assigned by parameter `-id`, it can be any string as long as not conflicts. The **first** in the list is used as the default mode.

| Mode Type  | Description  | Parameters |
| :------------ |:---------------| :-----|
| [idle]      | do nothing, normally used as default mode | `-id` modeId |
| [gyro] | enable/disable the gyroscope</br> on enter/exit       |    `-id` modeId|
| [speech]      | start/stop capturing audio input</br> on enter/exit  |  `-id` modeId</br> `-host` backend engine url, default: 127.0.0.1:2701</br>This backend uses a 128M model, there is also a 1.8GB docker image which consumes more memory but results in a better accuracy, can be installed with `docker run -d -p 2700:2700 alphacep/kaldi-en:latest` and set this param as: '-host 127.0.0.1:**2700**'. This model doesn't allow dynamic phrase_list, should only be used in sentence mode.</br>`-phrase` phrase id array that configured in **PhraseList** section.</br> &nbsp;&nbsp;&nbsp;&nbsp;e.g. '-phrase punctuation java cpp'</br>`-flushonexit` fire an **flush** event on mode exit to get recognition result quicker, see the action `[flush]` below |

**2. Mode Rule**

A Mode does very little, jobs are done by mode rules. There two types of rules:
- `trigger` -> `action`
**trigger**: one-time-event like button press, stick spinning, gyro rotating, speech text.
**action**: what it will do when above signal is triggered.
- `switch` -> `modifier`
**switch**: it can be turned on and off, when it's on, the **modifier** will be applied to the input signal.
e.g. `[switch] button -id R -> [boost] -speed 3` means when the button `R` is down, the cursor moves 3 times faster.

**Some examples:**

- `[trigger] stick -side Right -> [cursor] -speed 40`
Spin right stick to move mouse cursor.
- `[trigger] button -id R-SR  -> [hotkey] -keys t control alt` 
Press button R-SR to trigger hotkey "ctrl+alt+t".
- `[switch] button -id A  -> [prefix] -prefix "[camel] [title] "`
When button A is down, speech text is decorated to camel+title case, e.g.: "hello world again" -> "HelloWorldAgain".
- `[switch]  button -id R    -> [mode] -id MouseMode`
Switch to MouseMode by holding R, release R go get back to default mode.


**Note**: most parameters are set by single dash: `-text hello`, use double dash for boolean parameters: `--number=false`, and for array types: `-map a b c`, any special character should be wrapped with double quote, such as "-".

| trigger Type  | Description  | Parameters |
| :------------ |:---------------| :-----|
| [button]      | button down/up event | `-id` buttonId: </br>Y, X, B, A, R-SR, R-SL, R, ZR,</br> -, +, RStick, LStick, Home, Capture, </br>ChargingGrip, Down, Up, Right, Left,</br> L-SR, L-SL, L, ZL</br>Note: a double quote is required for the button "-" |
| [stick]      | stick spinning event | `-side` which Joy-Con, "Left" or "Right"|
| [gyro]      | when gyroscope is enabled | &nbsp;|
| [speech]   | when the voice is recognized and returned as text| &nbsp;|

| action Type  | Description  | Parameters  |
| :------------ |:---------| :-------------|
| [cursor]      | move mouse cursor  | `-speed` cursor move speed, float.</br> &gt;1 to increase, &lt;1 to decrease |
| [click]      |  mouse click  | `-button` "left", "center", "right", "wheelDown", "wheelUp", "wheelLeft", "wheelRight", default: "left"</br> `-double` is double click, default: false|
| [hotkey]   |  single key press or combination  | `-keys`  array of keys</br>e.g. "-keys enter" or "-keys t control alt"</br>Note: "t" first, then "control alt" </br> [key list](https://github.com/go-vgo/robotgo/blob/master/key.go#L205)|
| [notify]      | show a system notification  | `-title` title string</br>`-text` text body</br>`-icon` path of icon |
| [speech]      | execute a speech, words in a sentence can be executed in different ways, which can be configured in section **WordMapping**| `-number` convert number words to numeric digits, e.g. "twenty twenty two" -> "2022", default: true</br>`-nospace` remove space between words, for programming, default: true</br>`-typing` the word is typed if no other mapped executer handles it(like a hotkey), default: true</br>`-map` an array of group id in **WordMapping**, these mapping groups are used to handle this words, see that section for detail.</br>e.g. "-map desktop_hotkey golang python"|
| [speak]      |  used for complex task that cannot be done in a single action, works by simulating a speech text which will be handled by the above **[speech]** action| `-text` speech text to be executed |
| [flush]      |  this currently works by sending a chunk of zero data to speech engine, the engine may consider the zeroes as a long period of silence, hence it stops waiting for more voice input and returns result quicker. Only use this with limited phrase list, otherwise it can cause *stuck* behavior as it doesn't return result until next speech. | &nbsp;|
| [repeat]      |  repeat last action | &nbsp;|

| switch Type   | Description  | Parameters |
| :------------ |:---------------| :-----|
| [button]      | switched on when button down, off when button up | `-id` buttonId |
| [stick]      | switched on when stick moves to the edge, off when leaving that edge | `-side` "Left" or "Right"</br>`-dir` direction: Up/Down/Left/Right|

| modifier Type   | Description  | Parameters |
| :------------ |:---------------| :-----|
| [mode]      | switch to another mode | `-id` modeId |
| [boost]      | speed up/down cursor movement| `-multiplier` float number, &gt;1 to speed up, &lt;1 to slow down |
| [camel]</br>[title]</br>[snake]</br>[upper]      | convert speech text to different case by adding a prefix| &nbsp; |
| [prefix]      | add custom prefix to the speech text| `-prefix` prefix string</br>`-space` add a space between prefix and origin text, default: true|


**3. Phrase List**

A dictionary for limiting the speech model, only the words in the list are recognized. For example:
```
alphabet = ['a', 'b', 'c', ...'zed']
golang = ['package', 'switch', ...'']
```
Different groups can be used together for different speech modes. e.g. `-phrase alphabet punctuation application java lua`
If you found some words conflict a lot, like `4` and `for`, remove the `for` from the phrase list, map `4 loop` -> `for` , or `forever`->`for` in the mapping section below. Then the conflict is avoided:
- when you say `4`, it types `4`
- when you say `4 loop` or `forever`, it types `for`

**4. Word Mapping**

A sentence is splitted to many words and executed by different executors, which are registered in this WordMapping section, it can handle complex task like:
`run calc.exe`(shell) -> `delay 1 second`(delay) -> `type "1+1"`(typing) -> `press enter`(hotkey)

A list of executors:

| word executor Type   | Description  | Parameters |
| :------------ |:---------------| :-----|
| normal words     | if there is no tag([...]), it's simply a word replacement</br> Empty or special word should be wrapped with double quote "". | e.g. </br>`spring -> "fmt.Sprintf("` </br>`space -> " "`  |
| [hotkey]     | trigger a hotkey  | The key combination </br> [a full key list](https://github.com/go-vgo/robotgo/blob/master/key.go#L205)</br>e.g. `launch terminal -> [hotkey] t control alt` |
| [shell]     | execute a shell command  | command and arguments</br>e.g. `launch brave -> [shell] "brave-browser" "--no-sandbox" "www.test.com"` |
| [delay]     | delay some period | duration string</br>e.g. `sleep a while -> [delay] 1s`|
| [camel]</br>[title]</br>[upper]</br>[snake]</br>     |  case decorator | e.g. </br>`camel -> [camel]`</br>`elephant -> [camel] [title]` ThisIsElephant|
| [typing]    | the word will be typed if not handled by other executors |  |
| [repeat]    | repeat last speech |  |

Mappings are grouped and can be used together like `-map programming application go`

## TODO

- [ ] Auto change mode when switch between applications
- [ ] Show the speech text directly on screen



## Credits to
1. This is greatly inspired by [Talon Voice](https://talonvoice.com/ )
2. The awesome [dekuNukem/Nintendo_Switch_Reverse_Engineering](https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering)
3. All the Joy-Con protocol implementations:
[riking/joycon](https://github.com/riking/joycon "riking/joycon")
[wazho/ns-joycon](https://github.com/wazho/ns-joycon/ "wazho/ns-joycon")
[Davidobot/BetterJoy](https://github.com/Davidobot/BetterJoy "Davidobot/BetterJoy")
[tomayac/joy-con-webhid](https://github.com/tomayac/joy-con-webhid "tomayac/joy-con-webhid")
[looking-glass/joyconlib](https://github.com/looking-glass/joyconlib "looking-glass/joyconlib")

## License
MIT
