# Git Hub Tool. A tool for performing actions on  GitHub repos.

`ght` is a **read only** tool for extracting information from github.

Actions include:

- [List all repositories](#List-repos)
- [Show repoository details](#Show-repository-details)

Configure acccess by creating a [Github Personal API token](https://blog.github.com/2013-05-16-personal-api-tokens/) and saving it to `~/.ght`.  The token should have access to


## List repos

To list all of the repositories for an organisation ('MyOrg'), run the following command:

```
$ ght repos -o MyOrg
MyOrg/app1
MyOrg/app2
...
```

To list all of the repositories for a user ('sstephenson'), run the following command:

```
$ ght repos -u sstephenson
javan/blade
prototypejs/prototype
rails/actiontext
sstephenson/bashprof
sstephenson/bats
...
```

## Show repository details

To display repository summary, run the following, for example to show a summary of the [gogs project](https://github.com/gogs/gogs):

```
$ ght repo gogs/gogs
FullName           gogs/gogs
DefaultBranch      master
BranchProtection   None

Releases:
---------
Published           Tag          Author             Name
2018-09-17 03:57:54 v0.11.66     Unknwon            0.11.66
2018-06-05 11:44:57 v0.11.53     Unknwon            0.11.53
2018-03-31 19:36:26 v0.11.43     Unknwon            0.11.43
2017-11-23 08:52:48 v0.11.34     Unknwon            0.11.34
2017-11-20 07:38:06 v0.11.33     Unknwon            0.11.33
2017-08-16 10:23:43 v0.11.29     Unknwon            0.11.29
2017-06-11 07:51:09 v0.11.19     Unknwon            0.11.19
2017-04-06 05:28:12 v0.11.4      Unknwon            0.11.4
2017-04-04 12:34:24 v0.11        Unknwon            0.11
2017-03-28 13:52:16 v0.11rc      Unknwon            0.11 RC
2017-03-15 06:07:20 v0.10.18     Unknwon            0.10.18
2017-03-08 08:27:20 v0.10.8      Unknwon            0.10.8
2017-02-28 23:52:15 v0.10.1      Unknwon            0.10.1
2017-02-28 12:57:38 v0.10        Unknwon            0.10
2017-02-22 06:27:15 v0.10rc      Unknwon            0.10 RC
2017-02-11 22:08:31 v0.9.141     Unknwon            v0.9.141
2017-02-01 01:47:27 v0.9.128     Unknwon            v0.9.128
2016-12-24 16:04:42 v0.9.113     Unknwon            v0.9.113
2016-09-01 18:33:17 v0.9.97      Unknwon            v0.9.97
2016-08-11 07:19:17 v0.9.71      Unknwon            v0.9.71
```

The *Releases* section **ONLY** displays github releases, and **NOT** tags, for example compare the contents of [this repositories releases and tags](https://github.com/davidoram/gittest/releases) with the output: from `ght`:

```
$ ght repo davidoram/gittest
Full name :           davidoram/gittest
Default branch :      master
Branch protection (master), requires code review :  true
Branch protection (master), approval count :        1
Branch protection (master), branch must be up to date before merge :  true
Branch protection (master), status checks :  []

Releases:
---------
Status      Published           Tag          Author             Name
Pre-release 2018-11-07 08:46:01 v1.1.4       davidoram          Added support for Mac
Draft                           v1.1.2       davidoram          Added -p flag
Published   2017-07-10 16:29:40 v1.1.3       davidoram          The first release
```

Note that tags `v1.1.1` and `v1.1.2` are not returned in the list of releases.

Also note the branch protection settings are displayed for the default branch, in this case `master`.