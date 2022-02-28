---
id: contributing_plastic
title: Setup for Plastic SCM testing
sidebar_label: Plastic SCM testing
---

When changing the `segment_plastic.go` file, you may need to test your changes against an actual instance of
[Plastic SCM][plastic]. This doc should bring you up to speed with Plastic SCM.

In the [contributing doc][contributing] there is a section about [dev containers & codespaces][devcontainer].
You can setup Plastic SCM inside these as well.

## Server Setup

Here you can find the [official setup instructions][setup-instructions]. I'll describe it in short:

### Installation on Debian or in dev-container

First add the repo:

```bash
sudo apt-get update
sudo apt-get install -y apt-transport-https
echo "deb https://www.plasticscm.com/plasticrepo/stable/debian/ ./" | sudo tee /etc/apt/sources.list.d/plasticscm-stable.list
wget https://www.plasticscm.com/plasticrepo/stable/debian/Release.key -O - | sudo apt-key add -
sudo apt-get update
```

Then install the server: *this might throw an error at the end of the setup **see below***

```bash
sudo apt-get install plasticscm-server-core
```

This might show an error while configuring the installed package. In that case the server was nod registered as a service.
**Ignore it!**

### Server configuration

Configuring the server is done via:

```bash
cd /opt/plasticscm5/server/
sudo ./plasticd configure
```

You are asked 5 questions. Choose these options:

1. **1**: English
2. 8087 (default port, **just hit return**)
3. 8088 (default ssl port, **just hit return**)
4. **1**: NameWorkingMode (use local users and groups)
5. skip license token (**just hit return**)

**Congrats!** Your server is configured. You can find out more in the [official configuration instructions][server-config].

### Run Server

If your server installed without an error, it was correctly registered as a server and can be started via:

```bash
sudo service plasticd start
```

If not, you need to start it manually (for example inside the dev-container):

```bash
cd /opt/plasticscm5/server/
sudo ./plasticd start
```

This will lock the current shell until the server process finishes. You might need to open another terminal to continue.

Your Plastic SCM server should be started now.

## Client Setup

Plastic SCM comes, much like git, with a CLI (+ client UI \[optional\])

### Installation on Debian or in dev-container

These are the steps to install the **Plastic SCM CLI** on Debian or in the dev-container:

```bash
sudo apt-get install plasticscm-client-core
```

### Client configuration

To connect the client to the server and setup an account run:

```bash
clconfigureclient 
```

You are asked a few questions. Choose these options:

1. **1**: English
2. localhost (**just hit return**)
3. default port 8087 (**just hit return**)
4. No SSL (**just hit return**)
5. No Proxy (**just hit return**)

**Congrats!** Your client should now be connected to your server.

You can test if it worked and display some license info via:

```bash
cm li
```

## Testing stuff

Now to the fun part! The server is automatically setup to host a `default` repo with the branch `/main`.

The Plastic SCM CLI command is: `cm`

If you ever wonder what you can do with it call:

```bash
cm showcommands --all`
```

### Creating a local workspace

You need a local workspace to work with plastic:

```bash
cd ~
mkdir dev
cd dev
cm wk create workspace workspace rep:default
cd workspace
cm status
```

### Adding files

Start by creating local, private files

```bash
echo "test" > myfile.txt
cm status --all
```

Add the file to your local changes

```bash
cm add myfile.txt
cm status
```

**Test hint:** Both `Private` and `Added` files should be counted towards the `Added` property of the `plastic` segment.

### Commiting changes

After locally adding, changing, moving or deleting files you want to commit them to create a new changeset.
Run this command to commit all local changes:

```bash
cm status | cm ci . -c "my first commit"
```

### Undoing local changes

Just in case you don't want or can't commit your local changes, there is an undo command.
This will undo all local changes:

```bash
cm status | cm undo . 
```

### Changing, moving or deleting files

All these actions are done on the file level. You can run `cm status` to see your actions beeing tracked by plastic.
Use the commit method described above to commit your changes.

**Test hint:** All these changes should be counted by the designated property (`Modified`, `Moved`, `Deleted`)
of the `plastic` segment.

### Branching

Above the basics of handling the Plastic SCM client are described.
But you would want to dive deeper and use branches or labels and merge them.

#### Create a new branch

To create a new branch based on the latest changeset on branch `/main` call

```bash
cm br /main/new-branch
```

Hint: To list all branches use

```bash
cm find branches
```

#### Set a label to the current changeset

Your workspace will always reflect one specific changeset (see `cm status`). You can set a label on that changeset for
fast navigation or documentation purposes

```bash
cm label mk "BL0001"
```

Hint: To list all labels use

```bash
cm find labels
```

#### Switch your local workspace to a branch
  
To switch to a branch use

```bash
cm switch /main/new-branch
cm status
```

**Test Hint:** the branch name should be reflected in the `Selector` property of the `plastic` segment

#### Switch to a changeset

Each commit gets a unique changeset number. You can switch to these via

```bash
cm switch cs:1
```

**Test Hint:** the changeset should be reflected in the `Selector` property of the `plastic` segment

#### Switch to a label

You can also switch to a label via

```bash
cm switch BL00001
```

**Test Hint:** the label should be reflected in the `Selector` property of the `plastic` segment

#### Merge a branch

To merge a branch you have to switch to the *destination* branch of the merge. After that you can merge another branch via

```bash
cm switch /main
cm merge /main/new-branch --merge
cm status
```

Hint: This will only prepare the merge locally. You will have to commit the changes to complete the merge!

**Test Hint:** A pending merge should be reflected in the `MergePending` property of the `plastic` segment

#### Cherry-pick merge

While the merge above will merge all changes from a branch (and his parents), there is a cherry-pick merge,
which will merge only the changes of one single changeset

```bach
cm merge cs:8 --merge --cherrypicking
```

Hint: This will only prepare the merge locally. You will have to commit the changes to complete the merge!

**Test Hint:** A pending cherry-pick merge should be reflected in the `MergePending` property of the `plastic` segment

#### Merge conflicts

There are multiple causes for conflicts while merging

##### Evil Twin

This happens when a merge is performed where two files with the same name were added on both the source and destination branch.

```bash
cm br mk /main/sub-branch
cm switch /main/sub-branch
echo "1" > twin.txt
cm add twin.txt
cm ci twin.txt

cm switch /main
echo "2" > twin.txt
cm add twin.txt
cm ci twin.txt

cm merge /main/sub-branch --merge
```

Hint: this will prompt you to directly resolve the conflict

##### Changed on both sides

This happens when a merge is performed where a file was changed on both sides: source and destination

```bash
cm switch /main
echo "base" > file.txt
cm add file.txt
cm ci file.txt

cm br mk /main/test

echo "on main" > file.txt
cm ci file.txt

cm switch /main/test
echo "on test" > file.txt
cm ci file.txt

cm switch /main
cm merge /main/test --merge
```

Hint: this will try to open `gtkmergetool` which will fail inside the dev-container!

##### Changed vs. deleted file

This happens when a merge is performed where a file was modified on one side and deleted on the other side of the merge

```bash
cm switch /main
echo "base" > deleteme.txt
cm add deleteme.txt
cm ci deleteme.txt

cm br mk /main/del

rm deleteme.txt
cm ci --all

cm switch /main/del
echo "on del" > deleteme.txt
cm ci deleteme.txt

cm switch /main
cm merge /main/del --merge
```

Hint: This will prompt you to directly resolve the merge conflict

[plastic]: https://www.plasticscm.com/
[setup-instructions]: https://www.plasticscm.com/documentation/administration/plastic-scm-version-control-administrator-guide#Chapter3:PlasticSCMinstallation
[server-config]: https://www.plasticscm.com/documentation/administration/plastic-scm-version-control-administrator-guide#Serverconfiguration
[contributing]: https://github.com/JanDeDobbeleer/oh-my-posh/blob/main/CONTRIBUTING.md
[devcontainer]: https://github.com/JanDeDobbeleer/oh-my-posh/blob/main/CONTRIBUTING.md#codespaces--devcontainer-development-environment
