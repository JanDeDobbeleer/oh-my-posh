# Description: Post build script to compress the themes and generate SHA256 hashes for all files in the dist folder

# Compress all themes
$compress = @{
    Path             = "../themes/*.omp.*"
    CompressionLevel = "Fastest"
    DestinationPath  = "../src/dist/themes.zip"
}
Compress-Archive @compress

# Generate SHA256 hashes for all files in the dist folder
Get-ChildItem ./dist -Exclude *.yaml, *.sig | Get-Unique |
Foreach-Object {
    $zipHash = Get-FileHash $_.FullName -Algorithm SHA256
    $zipHash.Hash | Out-File -Encoding 'UTF8' "../src/dist/$($_.Name).sha256"
}
