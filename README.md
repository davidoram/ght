# Git Hub Tool. A tool for performing actions on  GitHub repos.

`ght` is a **read only** tool for extracting information from github.

Actions include:

- [List all repositories](#List-repos)
- [Show repoository details](#Show-repository-details)

Configure acccess by creating a [Github Personal API token](https://blog.github.com/2013-05-16-personal-api-tokens/) and saving it to `~/.ght`.  The token should have access to


## Install

Assumed you have go compiler installed.

```
make build install
```

## Usage

```
ght help
```

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

To see other options run `ght help repo`

## Show repository details

To display repository summary, run the following, for example to show a summary of the [gittest project](https://github.com/davidoram/gittest):

```
$ ght repo davidoram/gittest
Repository
----------
Full name:    	davidoram/gittest
Default branch	master
URL           	https://github.com/davidoram/gittest

Branch Protection
-----------------
Branch:                  	master
- approving review:      	true
- approving review count:	1
- approving review count:	1
- status check:          	true
- status check contexts: 	[]

Releases
--------
Tag:
Release status:	Draft
Published at:  	0001-01-01 11:30:00
Author:        	davidoram
URL:           	https://github.com/davidoram/gittest/releases/tag/untagged-487af99a8c4714a10c5e
Name:          	Added -p flag


Tag:           	v1.1.4
Release status:	Pre-release
Published at:  	2018-11-07 08:46:01
Author:        	davidoram
URL:           	https://github.com/davidoram/gittest/releases/tag/v1.1.4
Name:          	Added support for Mac


Tag:           	v1.1.3
Release status:	Published
Published at:  	2017-07-10 16:29:40
Author:        	davidoram
URL:           	https://github.com/davidoram/gittest/releases/tag/v1.1.3
Name:          	The first release


Tags
----
Tag   	Sha
v1.1.1	3b0237e261d565e2c01de15e8a703a989cdf1a62
v1.1.2	1f6fb3fd0510b557f8519e5cda58beb9705a82f1
v1.1.3	7e39483974054e74b07f690742ac0fda8fe4e72d
v1.1.4	166418892de06bd1671bd0d8fa7897abe64bd22d
```

### Notes:

- The *Releases* section **ONLY** displays github releases, and **NOT** standard tags. Eg: compare the following example:
  - Tags https://github.com/davidoram/gittest/tags.
  - Release https://github.com/davidoram/gittest/releases
- Branch protection settings are displayed for the all branches, in this example `master` has some protection settings enabled.
- Run `ght help repo` to see other options


## Show changelog

To display a projects github releases in the form of a traditional [changelog](https://en.wikipedia.org/wiki/Changelog#Format), run the following, for example to show the change log for [gittest project](https://github.com/davidoram/gittest):

```
$ ght repo davidoram/gittest -changelog
# Change Log

## Untagged

- Added -p flag
  - **Features**
 
 - Added the -p flag
 
 **Other**
 
 - Added support for RasbperryPi

## v1.1.4

- Added support for Mac
  - **Other**
 
 - Added support for Mac

## v1.1.3

- The first release
  - **Bug Fixes**
 
 - Fixed the bug that broke stuff
 
 **Features**
 
 - Added the -l flag
 
 **Improvements**
 
 - Speed improved x2
 
 **Other**
 
 - Added support for Windows


```

### Notes:

- Uses the same logic as the [Show repository details](#Show_repository_details)
- Pass the `-color` flag to add a sprinking of color, which makes it a bit easier to can and read
