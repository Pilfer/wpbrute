## Author enumeration #####

Enumerate those authors, yo.  

### REST API  
- Visit `/wp-json/wp/v2/users?per_page=100&offset=0`
- If that returns more than 100 users, loop until we get all users.

### ?author={id} enumeration  

Visit ?author=n for n in range 1 to 10 for all sites.  
For sites where range 1-10 returned at least 1 author ID, continue enumerating author until 10 consecutive 404s.

Success condition: ?author=1 redirects to /author/username  


### Google Search on Author Subfolder  

Use ValueSERP or equivalent and do search for `site:domain.com/author/ -inurl:page`.  
Extract `domain.com/author/username` from urls (stripping any `?query=params` and stopping at `/`).

e.g. `example.com/author/username?ok=1` -> `username`  
`example.com/author/username/` -> `username`  

Set parameter to receive 100 results per search page. If first result contains more than 100 results, continue interating down pages until finished.  


### Author-Sitemap.xml Parse

Extract author urls from sitemap.

**Examples**:

```http://thinkific.com/author-sitemap.xml
http://brookings.edu/author-sitemap.xml
http://realtor.com/author-sitemap.xml
http://zenfolio.com/author-sitemap.xml
http://marketingland.com/author-sitemap.xml
http://snowplowanalytics.com/author-sitemap.xml
http://pardot.com/author-sitemap.xml
http://smallbiztrends.com/author-sitemap.xml
http://digiday.com/author-sitemap.xml
http://jextensions.com/author-sitemap.xml
http://webriti.com/author-sitemap.xml
http://networkforgood.com/author-sitemap.xml
http://akc.org/author-sitemap.xml
http://gigaom.com/author-sitemap.xml
http://pro.photo/author-sitemap.xml
http://lever.co/author-sitemap.xml
http://earthlink.net/author-sitemap.xml
http://blubrry.com/author-sitemap.xml
http://affordable-papers.net/author-sitemap.xml
http://nobelprize.org/author-sitemap.xml
```
