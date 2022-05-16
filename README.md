# WPBrute - An Offensive Security Tool for Wordpress  


**Disclaimer**: *The intended purpose of this software is for the betterment of the internet. This software is only meant to serve educational _or_ professional purposes where legal and against consenting websites. Any usage outside of the aforementioned scope is forbidden.*  

---  


**Wordpress** is a content management system that people mistakenly use as a blog platform sometimes */s*. Like Internet Explorer, nobody likes using it but we're stuck with it for some reason. I don't know why.

**WPBrute** is an open-source offensive security tool that was built to enable security professionals to check their Wordpress install hardening, as well as emulate and defend against known login-based attack vectors (read: `oops we don't rate-limit wp-login.php and we left xmlrpc.php open for the whole world to see`).  

This tool can handle multiple targets if required - ideal for engagements where the target operates multiple instances of Wordpress throughout the organization.  


### Usage  

See `./wpbrute -h`.  

```
NAME:
   wpbrute - A bespoke security tool for redteamers to test wordpress credentials on a variety of targets. See the 'help' command for usage instructions.

USAGE:
   wpbrute [global options] command [command options] [arguments...]

COMMANDS:
   load     
   report   
   recon    
   brute    
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value  Config file (default: "config.yml")
   --migrate       Whether or not to migrate the database (default: false)
   --help, -h      show help (default: false)
```


### Features  

**Note**: Some of these features were lost in the great laptop-death of 2021. I don't push as often as I should. Sorry. Feel free to reimplement at your leisure.  

- Persists results/etc to a database (postgres and sqlite by default)  
  - Configurable - see `config-example.yaml` and `./pkg/config.go`.  
- ~~Simple web interface~~ **Lost to broken laptop**    
- Multi-site support  
- HTTP/HTTPS/SOCKS5 Proxy Support  
- Test and identify compromised credentials  
- Test logins from a list against:  
    - `wp-login.php` with concurrency  
    - `xmlrpc.php` with concurrency  
- ~~Check for `/.wp-config.php.swp`~~ **Lost to broken laptop**  
    - ~~If admins use vim to edit these settings then it might be possible to snag the swap file. The analyst performing the audit can spearphish the security team and let them know the password showed up in a leak and that it needs to be changed, all while polling for the swap file over HTTP(s).~~  
- ~~Source and/or enumerate author names through:~~ **Lost to broken laptop, sorry!**  
    - ~~SERP dorks (requires API key for most services)~~  
    - ~~Author enumeration (via `HEAD ?author=${n}`)~~  
    - ~~`/wp-json/` API~~  
    - ~~`author_sitemap.xml` parsing~~  

#### Issues  

I'm not supporting this trash code lol. I'll check out PRs if you have any, however the enterprising individual might see fit to just implement the code in their preferred language/stack and maintain full control over it. This repo exists largely as a proof-of-concept.  
  

#### Backlog    
- [X] Load password list(s) to the database  
- [X] Permutate possible username/password combinations  
  - [X] Explode and persist `bob.smith@example.com` as `bob.smith@example.com`, `bob.smith`, and `bob` for possible username combinations with the associated password  
- [X] Generate default username/password combinations  
- [X] Check + store if site is using HTTPS   
- [X] Check + store if site redirects to a subdomain or another host   
- [X] Check + store if `/xmlrpc.php` is present and enabled   
- [X] Check + store if `/wp-login.php` is present and not filtered   
- [ ] Check + store if `/wp-admin/*` is present and not filtered   
- [ ] Check for `/.wp-config.php.swp` on a loop for `n` duration of time  
- [ ] Source authors from SERPs.  
- [ ] Source authors from author enumeration.  

