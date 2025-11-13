---
next:
  text: 'Modules'
  link: '/docs/modules'
---
# Configuration

Configuration file is divided into 4 sections:
- `bar`: used for general bar configuration
- `left`: left anchored modules
- `middle`: centered modules
- `right`: right anchored modules

# `bar`
Has three options:
- `truncate_priority`
- `enable_ellipsis`
- `ellipsis`
## `truncate_priority`
Sets content priority on overlap between anchored modules.

Default:
```yaml
truncate_priority:
  - right
  - left
  - middle
```

For example, in the config
```yaml
left:
  - title
middle:
  - clock
```

if the content of the title reaches clock then by default clock will be under the title.

one set of anchor can be given priority over the other with this

## `enable_ellipsis`
Enables adding ellipsis inside truncated content

Adds ellipsis at the point of overlap. The ellipsis uses the space of the anchor being truncated.

Default:
```yaml
enable_ellipsis: true
```

## `ellipsis`
Sets the ellipsis content.

Default:
```yaml
ellipsis: "…"
```

# Modules

There are **16** modules currently:
- `backlight`
- `battery`
- `bluetooth`
- `clock`
- `cpu`
- `custom`
- `disk`
- `idleInhibitor`
- `locale`
- `mpris`
- `ram`
- `title`
- `tray`
- `volume`
- `wifi`
- `ws`


Each module must be anchored in one of **3** ways:
- `left`
- `middle`
- `right`

like this in config file:
```yaml
left:
  - ws
  - title
middle:
  - clock
right:
  - tray
  - sep
  - disk
```

The modules are listed in a way such that top maps to left and bottom maps to right.\
So in the example above, tray will be the leftmost module in the right anchor and disk be the rightmost

Each modules has some configuration options, but some common configuration options are available in most modules like:
- `fg`
- `bg`
- `format`
- `cursor`
- `onmouse`

## `fg`
Set foreground color.
\
Changes color for text, icons, etc

Example:
```yaml
fg: aliceblue
```

### Colors
Colors can set using **4** different methods:
- [CSS Named Colors](https://developer.mozilla.org/en-US/docs/Web/CSS/named-color): Set the name directly like `fg: rebeccapurple`
- Hex Codes: `fg: #1A553B` or `fg: #FFF`
- RGB Codes: `fg: rgb(234,98,102)`
- Predefined Variables: `fg: @urgent`, `fg: @good`, `fg: @color112` 

## `bg`
Set background color.

Example:
```yaml
bg: saddlebrown
```

Refer [Colors](#colors) for how to set colors

## `format`
Sets display format for the module

Apart from `clock`, most modules uses go's [text/template format syntax](https://pkg.go.dev/text/template)

For basic configuration it is enough to know, each module has its set of keywords, like `backlight` may have:
- `Icon`: icon
- `Percent`: percent of backlight
- `Now`: current brightness units
- `Max`: maximum brightness units

Now, these can be escaped using the syntax <code>&#123;&#123;.Keyword}}</code> \
Examples:
```yaml
format: "{{.Icon}} {{.Percent}}%" 
```

```yaml
format: "hallo: {{.Icon}}: {{.Now}}/{{.Max}}"
```

templates are really powerful and you can do a bunch of cool stuff with it. Check out the link above to know more.

## `cursor`
Sets cursor shown while hovering over the module.

```yaml
cursor: pointer
```

Available cursor shapes can be referred from [here](https://developer.mozilla.org/en-US/docs/Web/CSS/cursor)

## onmouse
Sets dynamic behavior on mouse interaction

Supports setting behaviour on 7 types of mouse interaction:
- `left`: Left click
- `right`: Right click
- `middle`: Middle click
- `wheel-up`: Scroll up (touchpad or mouse)
- `wheel-down`: Scroll down
- `wheel-left`: Scroll left
- `wheel-right`: Scroll right
- `hover`: Hover over the module, only triggers when you enter the module space

Each type of interaction has **3** actions it can perform when triggered:
- `run`: Execute a command
- `notify`: Send a desktop notification
- `config`: Change config dynamically

### `run`
Execute a command

Example:
```yaml
onmouse:
  left:
    run: "pavucontrol"
```

### `notify`
Send a desktop notification

Example:
```yaml
onmouse:
  middle:
    notify: "Hey I am pawbar."
```

### `config`
List of configs for this module to cycle through each time this type of interaction is triggered.

Example:
```yaml
onmouse:
  left:
    config:
      - fg: blue # conf 1
        bg: green
      - fg: red # conf 2
        bg: white
```

On each left click, this will cycle through the base config, conf 1, conf 2

Each module will have its set of allowed config options in onmouse interactions, but most will have atleast `fg`, `bg`, `format`, `cursor`

