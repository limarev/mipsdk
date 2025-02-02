# mipsdk-scraper

Application parses html located at `url` and is looking for \"downloadFile\" json field. This field contains json array. Each item of the array is a mip binary descriptor. This descriptor contains URL where mip binary can be downloaded from. Application downloads the mip binaries in parallel to specified `dir` if scraping was successful.
```
% ./mipsdk-scraper -h                    
Usage of ./mipsdk-scraper:
  -dir string
        download dir (default ".")
  -timeout int
        downloading timeout per file in seconds (default 600)
  -url string
        url for scraping (default "https://aka.ms/mipsdkbins")
  -version-only
        no downloading actually happens, returns mipsdk binaries version if found
```
`url` for mipsdk-scraper can be found at [release history section](https://learn.microsoft.com/en-us/information-protection/develop/version-release-history#release-history).

go.yml github action creates release with the mip sdk binaries uploaded and tag corresponding mipsdk version. Press `Run workflow` to  [run action](https://github.com/limarev/mipsdk-scraper/actions/workflows/go.yml).
