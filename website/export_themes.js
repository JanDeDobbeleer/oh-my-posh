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
  return theme.endsWith('.omp.json') || theme.endsWith('.omp.toml') || theme.endsWith('.omp.yaml');
}

async function* asyncPool(concurrency, iterable, iteratorFn) {
  // https://github.com/rxaviers/async-pool/blob/master/lib/es9.js
  const executing = new Set();
  async function consume() {
    const [promise, value] = await Promise.race(executing);
    executing.delete(promise);
    return value;
  }
  for (const item of iterable) {
    // Wrap iteratorFn() in an async fn to ensure we get a promise.
    // Then expose such promise, so it's possible to later reference and
    // remove it from the executing pool.
    const promise = (async () => await iteratorFn(item, iterable))().then(
      value => [promise, value]
    );
    executing.add(promise);
    if (executing.size >= concurrency) {
      yield await consume();
    }
  }
  while (executing.size) {
    yield await consume();
  }
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

async function exportTheme(theme) {
  if (!isValidTheme(theme)) {
    return;
  }

  const configPath = path.join(themesConfigDir, theme);

  const config = themeConfigOverrrides.get(theme) || newThemeConfig();

  const themeName = theme.slice(0, -9);
  const image = themeName + '.png';

  let poshCommand = `oh-my-posh config export image --config=${configPath} --output=${image}`;
  poshCommand += ` --background-color=${config.bgColor}`;
  if (config.author !== '') {
    poshCommand += ` --author="${config.author}"`;
  }

  const { _, stderr } = await exec(poshCommand);

  if (stderr !== '') {
    console.error(`Unable to create image for ${theme}, please try manually`);
    return;
  }

  console.info(`Exported ${theme} to ${image}`);

  const toPath = path.join(themesStaticDir, image);

  await fs.promises.rename(image, toPath);

  const themeData = `
### [${themeName}]

[![${themeName}](/img/themes/${themeName}.png)][${themeName}]
`;
  const link = `[${themeName}]: https://github.com/JanDeDobbeleer/oh-my-posh/blob/main/themes/${theme} '${themeName}'\n`;
  return {themeData, link};
}

(async () => {
  const themes = await fs.promises.readdir(themesConfigDir);
  const links = [];
  for await (const result of asyncPool(8, themes, exportTheme)) {
    if (!result) { // invalid theme or unable to create image
      continue;
    }
    const {themeData, link} = result;
    await fs.promises.appendFile('./docs/themes.md', themeData);
    links.push(link);
  }

  await fs.promises.appendFile('./docs/themes.md', '\n');

  for (const link of links) {
    await fs.promises.appendFile('./docs/themes.md', link);
  }
})();
