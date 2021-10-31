//jshint esversion:8
//jshint node:true
const fs = require('fs');
const path = require('path');
const util = require('util');
const exec = util.promisify(require('child_process').exec);

const themesConfigDir = "./../themes";
const themesStaticDir = "./static/img/themes";

function newThemeConfig(rpromptOffset = 40, cursorPadding = 30, author = "", bgColor = "#151515") {
  var config = {
    rpromptOffset: rpromptOffset,
    cursorPadding: cursorPadding,
    author: author,
    bgColor: bgColor
  };
  return config;
}

let themeConfigOverrrides = new Map();
themeConfigOverrrides.set('agnoster.omp.json', newThemeConfig(40, 40));
themeConfigOverrrides.set('agnosterplus.omp.json', newThemeConfig(80));
themeConfigOverrrides.set('amro.omp.json', newThemeConfig(40, 100, 'AmRo', '#1C2029'));
themeConfigOverrrides.set('avit.omp.json', newThemeConfig(40, 80));
themeConfigOverrrides.set('blueish.omp.json', newThemeConfig(40, 100));
themeConfigOverrrides.set('cert.omp.json', newThemeConfig(40, 50));
themeConfigOverrrides.set('cinnamon.omp.json', newThemeConfig(40, 80));
themeConfigOverrrides.set('craver.omp.json', newThemeConfig(40, 80, 'Nick Craver', '#282c34'));
themeConfigOverrrides.set('darkblood.omp.json', newThemeConfig(40, 40));
themeConfigOverrrides.set('honukai.omp.json', newThemeConfig(20));
themeConfigOverrrides.set('hotstick.minimal.omp.json', newThemeConfig(40, 10));
themeConfigOverrrides.set('hunk.omp.json', newThemeConfig(40, 15, 'Paris Qian'));
themeConfigOverrrides.set('huvix.omp.json', newThemeConfig(40, 70));
themeConfigOverrrides.set('jandedobbeleer.omp.json', newThemeConfig(40, 15));
themeConfigOverrrides.set('lambda.omp.json', newThemeConfig(40, 40));
themeConfigOverrrides.set('marcduiker.omp.json', newThemeConfig(0, 40));
themeConfigOverrrides.set('material.omp.json', newThemeConfig(40, 40));
themeConfigOverrrides.set('microverse-power.omp.json', newThemeConfig(40, 100));
themeConfigOverrrides.set('negligible.omp.json', newThemeConfig(10));
themeConfigOverrrides.set('night-owl.omp.json', newThemeConfig(40, 0, 'Mr-Vipi', '#011627'));
themeConfigOverrrides.set('paradox.omp.json', newThemeConfig(40, 100));
themeConfigOverrrides.set('powerlevel10k_classic.omp.json', newThemeConfig(10));
themeConfigOverrrides.set('powerlevel10k_lean.omp.json', newThemeConfig(80));
themeConfigOverrrides.set('powerline.omp.json', newThemeConfig(40, 40));
themeConfigOverrrides.set('pure.omp.json', newThemeConfig(40, 80));
themeConfigOverrrides.set('remk.omp.json', newThemeConfig(40, 40));
themeConfigOverrrides.set('robbyrussel.omp.json', newThemeConfig(40, 40));
themeConfigOverrrides.set('slim.omp.json', newThemeConfig(10, 80));
themeConfigOverrrides.set('slimfat.omp.json', newThemeConfig(10, 93));
themeConfigOverrrides.set('space.omp.json', newThemeConfig(40, 40));
themeConfigOverrrides.set('spaceship.omp.json', newThemeConfig(40, 40));
themeConfigOverrrides.set('star.omp.json', newThemeConfig(40, 70));
themeConfigOverrrides.set('stelbent.minimal.omp.json', newThemeConfig(70));
themeConfigOverrrides.set('tonybaloney.omp.json', newThemeConfig(0,40));
themeConfigOverrrides.set('unicorn.omp.json', newThemeConfig(0,40));
themeConfigOverrrides.set('ys.omp.json', newThemeConfig(40, 100));
themeConfigOverrrides.set('zash.omp.json', newThemeConfig(40, 40));

(async () => {
  const themes = await fs.promises.readdir(themesConfigDir);
  let links = new Array();

  for (const theme of themes) {
    if (!theme.endsWith('.omp.json')) {
      continue;
    }
    const configPath = path.join(themesConfigDir, theme);

    let config = newThemeConfig();
    if (themeConfigOverrrides.has(theme)) {
      config = themeConfigOverrrides.get(theme);
    }

    let poshCommand = `oh-my-posh --config=${configPath} --shell shell --export-png`;
    poshCommand += ` --rprompt-offset=${config.rpromptOffset}`;
    poshCommand += ` --cursor-padding=${config.cursorPadding}`;
    poshCommand += ` --bg-color=${config.bgColor}`;
    if (config.author !== '') {
      poshCommand += ` --author="${config.author}"`;
    }

    const { _, stderr } = await exec(poshCommand);

    if (stderr !== '') {
      console.error(`Unable to create image for ${theme}, please try manually`);
      continue;
    }

    const image = theme.replace('.omp.json', '.png');
    const toPath = path.join(themesStaticDir, image);

    await fs.promises.rename(image, toPath);

    const themeName = theme.replace('.omp.json', '');

    const themeData = `
### [${themeName}]

[![${themeName}](/img/themes/${themeName}.png)][${themeName}]
`;

    await fs.promises.appendFile('./docs/themes.md', themeData);

    links.push(`[${themeName}]: https://github.com/JanDeDobbeleer/oh-my-posh/blob/main/themes/${theme} '${themeName}'\n`);
  }

  for (const link of links) {
    await fs.promises.appendFile('./docs/themes.md', link);
  }
})();
