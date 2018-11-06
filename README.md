Git Hub Tool. A tool for performing actions on  GitHub repos.

Actions include:

- List repos

Configure acccess by creating a [Github Personal API token](https://blog.github.com/2013-05-16-personal-api-tokens/) and saving it to `~/.ght`.  The token should have access to


## List repos

List repos as follows

```
$ ght repos -o MyOrg
MyOrg/app1
MyOrg/app2
...
```