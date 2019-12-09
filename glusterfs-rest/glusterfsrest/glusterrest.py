# -*- coding: utf-8 -*-
"""
    glusterrest.py

    :copyright: (c) 2014 by Aravinda VK
    :license: MIT, see LICENSE for more details.
"""

import getpass
import sys
import os
from glusterfsrest import cliargs, users
from glusterfsrest.config import PORT_FILE, GROUPS, USERS_DB


def validate_groupname(grp):
    if grp not in GROUPS:
        sys.stderr.write("Invalid group %s\nAvailable groups: %s" %
                         (grp, ",".join(GROUPS)))
        sys.exit(1)


def show_users():
    usersdata = users.get()
    sys.stdout.write("%20s %10s\n" % ("User", "Group"))
    for u in usersdata:
        sys.stdout.write("%20s %10s\n" % (u[0], u[1]))


def show_config():
    sys.stdout.write("%10s %10s\n" % ("Config", "Value"))
    port = 9000
    with open(PORT_FILE) as f:
        port = f.read()
    sys.stdout.write("%10s %10s\n" % ("Port", port))


def show_groups():
    sys.stdout.write("Available Groups:\n----------------\n")
    for g in GROUPS:
        sys.stdout.write("%s\n" % g)


def set_port(value):
    with open(PORT_FILE, "w") as f:
        f.write(str(value))

    os.chmod(PORT_FILE, 644)
    sys.stdout.write("Port updated successfully, "
                     "Restart glusterrestd to use the latest port\n")
    return 0


def useradd(username, password, groupname):
    return users.useradd(username, password, groupname)


def usermod(username, groupname):
    return users.usermod(username, groupname)


def userdel(username):
    return users.userdel(username)


def passwd(username, password):
    return users.passwd(username, password)


def get_password():
    passwd = getpass.getpass('Password: ')
    confirm_passwd = getpass.getpass('Confirm Password: ')
    if passwd != confirm_passwd:
        sys.stderr.write("Pwd didn't match\n")
        sys.exit(1)

    if not passwd:
        sys.stderr.write("Invalid pwd\n")
        sys.exit(1)

    return passwd


def main_i():
    if os.getuid() != 0:
        sys.stderr.write("Only root can run this\n")
        sys.exit(1)

    args = cliargs.get()
    users.connect()
    ret = 0
    if args.subcommand == 'install':
        ret = users.install()
        os.chmod(USERS_DB, 644)
        os.chmod(PORT_FILE, 644)
    elif args.subcommand == 'reinstall':
        ret = users.reinstall()
        os.chmod(USERS_DB, 644)
        os.chmod(PORT_FILE, 644)
    elif args.subcommand == 'show':
        ret = globals()["show_%s" % args.option]()
    elif args.subcommand == 'port':
        ret = set_port(args.port)
    elif args.subcommand == 'useradd':
        if users.exists(args.username):
            sys.stderr.write('User %s already exist\n' % args.username)
            sys.exit(1)

        validate_groupname(args.group.lower().strip())

        if args.password == '':
            args.password = get_password()
        ret = useradd(args.username, args.password, args.group.lower().strip())
    elif args.subcommand == 'usermod':
        if not users.exists(args.username):
            sys.stderr.write('User %s does not exist\n' % args.username)
            sys.exit(1)
            validate_groupname(args.group.lower().strip())
        ret = usermod(args.username, args.group.lower().strip())
    elif args.subcommand == 'userdel':
        if not users.exists(args.username):
            sys.stderr.write('User %s does not exist\n' % args.username)
            sys.exit(1)

        ret = userdel(args.username)
    elif args.subcommand == 'passwd':
        if not users.exists(args.username):
            sys.stderr.write('User %s does not exist\n' % args.username)
            sys.exit(1)

        if args.password == '':
            args.password = get_password()
        ret = usermod(args.username, args.password)

    sys.exit(0 if ret else 1)


def main():
    try:
        main_i()
    except KeyboardInterrupt:
        sys.exit(1)


if __name__ == '__main__':
    main()
