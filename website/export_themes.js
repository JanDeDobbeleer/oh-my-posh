//jshint esversion:8
//jshint node:true
const fs = require('fs');
const path = require('path');
const util = require('util');
const exec = util.promisify(require('child_process').exec);

const themesConfigDir = "./../themes";
const themesStaticDir = "./static/img/themes";

function newThemeConfig(author = "", bgColor = "#151515") {
  var config = {
    author: author,
    bgColor: bgColor
  };
  return config;
}

function isValidTheme(theme) {
  return theme.endsWith('.omp.json') || theme.endsWith('.omp.toml') || theme.endsWith('.omp.yaml')
}

let themeConfigOverrrides = new Map();
themeConfigOverrrides.set('amro.omp.json', newThemeConfig('AmRo', '#1C2029'));
themeConfigOverrrides.set('chips.omp.json', newThemeConfig('CodexLink | v1.2.4, Single Width (07/11/2023) | https://github.com/CodexLink/chips.omp.json'));
themeConfigOverrrides.set('craver.omp.json', newThemeConfig('Nick Craver', '#282c34'));
themeConfigOverrrides.set('hunk.omp.json', newThemeConfig('Paris Qian'));
themeConfigOverrrides.set('kushal.omp.json', newThemeConfig('Kushal-Chandar'));
themeConfigOverrrides.set('night-owl.omp.json', newThemeConfig('Mr-Vipi', '#011627'));
themeConfigOverrrides.set('quick-term.omp.json', newThemeConfig('SokLay'))
themeConfigOverrrides.set('catppuccin.omp.json', newThemeConfig('IrwinJuice', '#24273A'));
themeConfigOverrrides.set('catppuccin_latte.omp.json', newThemeConfig('IrwinJuice', '#EFF1F5'));
themeConfigOverrrides.set('catppuccin_frappe.omp.json', newThemeConfig('IrwinJuice', '#303446'));
themeConfigOverrrides.set('catppuccin_macchiato.omp.json', newThemeConfig('IrwinJuice', '#24273A'));
themeConfigOverrrides.set('catppuccin_mocha.omp.json', newThemeConfig('IrwinJuice', '#1E1E2E'));

(async () => {
  const themes = await fs.promises.readdir(themesConfigDir);
  let links = new Array();

  for (const theme of themes) {
    if (!isValidTheme(theme)) {
      continue;
    }
    const configPath = path.join(themesConfigDir, theme);

    let config = newThemeConfig();
    if (themeConfigOverrrides.has(theme)) {
      config = themeConfigOverrrides.get(theme);
    }

    let poshCommand = `oh-my-posh config export image --config=${configPath}`;
    poshCommand += ` --background-color=${config.bgColor}`;
    if (config.author !== '') {
      poshCommand += ` --author="${config.author}"`;
    }

    const { _, stderr } = await exec(poshCommand);

    if (stderr !== '') {
      console.error(`Unable to create image for ${theme}, please try manually`);
      continue;
    }

    console.info(`Exported ${theme}`);

    const themeName = theme.slice(0, -9);
    const image = themeName + '.png';
    const toPath = path.join(themesStaticDir, image);

    await fs.promises.rename(image, toPath);

    const themeData = `
### [${themeName}]

[![${themeName}](/img/themes/${themeName}.png)][${themeName}]
`;

    await fs.promises.appendFile('./docs/themes.md', themeData);

    links.push(`[${themeName}]: https://github.com/JanDeDobbeleer/oh-my-posh/blob/main/themes/${theme} '${themeName}'\n`);
  }

  await fs.promises.appendFile('./docs/themes.md', '\n');

  for (const link of links) {
    await fs.promises.appendFile('./docs/themes.md', link);
  }
})();
