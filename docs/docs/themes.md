---
id: themes
title: Themes
sidebar_label: 🎨 Themes
---

Oh My Posh comes with many themes included out-of-the-box. Below are some screenshots of the more common themes.
For the full updated list of themes, [view the themes][themes] in Github.  If you are using PowerShell, you can
display every available theme using the following PowerShell cmdlet.

```powershell
Get-PoshThemes
```

Once you're ready to swap to a theme, follow the steps described in [🚀Installation/Replace your existing prompt][replace-you-existing-prompt].

Themes with `minimal` in their names do not require a Nerd Font. Read about [🆎Fonts][fonts] for more information.

To set a random theme everytime you open terminal, you can use the following:

###### Windows PowerShell

This assumes you have installed oh-my-posh using winget.

```PowerShell
$theme = Get-ChildItem $env:UserProfile\AppData\Local\Programs\oh-my-posh\themes\ | Get-Random
echo $theme.name
oh-my-posh --init --shell pwsh --config $theme.FullName | Invoke-Expression
```
[Source](https://gist.github.com/MRDGH2821/69aa46f5c52daecb140c7f7ccdecf969)

###### Windows Command Prompt

This assumes you have installed [clink](https://mridgers.github.io/clink/) and oh-my-posh via winget method.

```lua
math.randomseed(os.time())
math.random()
math.random()
math.random()
function math.randomchoice(t) --Selects a random item from a table
    local keys = {}
    for key, value in pairs(t) do
        keys[#keys + 1] = key --Store keys in another table
    end
    index = keys[math.random(1, #keys)]
    return t[index]
end

-- Lua implementation of PHP scandir function
local i, themes, popen = 0, {}, io.popen
local pfile = popen('dir "%userprofile%\\AppData\\Local\\Programs\\oh-my-posh\\themes\\" /b')
for filename in pfile:lines() do
    i = i + 1
    themes[i] = filename
end
pfile:close()

local theme = math.randomchoice(themes)
local query =
    string.format(
    "oh-my-posh --config=%%userprofile%%\\AppData\\Local\\Programs\\oh-my-posh\\themes\\%s --init --shell cmd",
    theme
)
print(theme) --This line can be removed if you don't want to know which theme is loaded.
load(io.popen(query):read("*a"))()
```
[Source](https://gist.github.com/MRDGH2821/a05aa96724484e3bce55c3c66855b933)

[themes]: https://github.com/JanDeDobbeleer/oh-my-posh/tree/main/themes
[fonts]: /docs/config-fonts
[replace-you-existing-prompt]: /docs/windows#override-the-theme-settings

<!-- Do not change the content below, themes are rendered automatically -->
