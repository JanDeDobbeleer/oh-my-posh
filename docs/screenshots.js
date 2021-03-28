const ansiToImage = require('ansi-to-image')
const util = require('util');
const exec = util.promisify(require('child_process').exec);

async function main() {
  const { stdout, stderr } = await exec('oh-my-posh --config ../themes/jandedobbeleer.omp.json --shell uni --pwd ~/Projects/oh-my-posh');
  if (stderr) {
    console.error(`error: ${stderr}`);
    return;
  }
  ansiToImage(stdout, {
    filename: 'static/img/themes/test.png',
    type: 'png',
    scale: 2,
    fontFamily: 'VictorMono Nerd Font Mono',
    fontSize: 13,
    lineHeight: 18,
    paddingTop: 10,
    paddingLeft: 10,
    paddingBottom: 10,
    paddingRight: 10,
    colors: './one-dark.itermcolors',
  });
}

main();
