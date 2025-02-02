# mipsdk-scraper

Application scrapes json from `url` with mip sdk binaries urls and then downloads the binaries in parallel to specified `dir`.
```
% ./mipsdk-scraper -h                    
Usage of ./mipsdk-scraper:
  -dir string
        download dir (default ".")
  -timeout int
        file downloading timeout in seconds (default 600)
  -url string
        url for scraping (default "https://aka.ms/mipsdkbins")
  -version-only
        no downloading actually happens, returns mipsdk binaries version if found
```

go.yml github action creates release with the mip sdk binaries uploaded. Press `Run workflow` to  [run action](https://github.com/limarev/mipsdk-scraper/actions/workflows/go.yml).
