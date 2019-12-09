# -*- coding: utf-8 -*-
"""
    cliargs.py

    :copyright: (c) 2014 by Aravinda VK
    :license: MIT, see LICENSE for more details.
"""

from argparse import RawDescriptionHelpFormatter, ArgumentParser

PROG_DESCRIPTION = """
GlusterFS REST API Client
"""


def get():
    parser = ArgumentParser(formatter_class=RawDescriptionHelpFormatter,
                            description=PROG_DESCRIPTION)

    subparser = parser.add_subparsers(dest='subcommand',
                                      title='Available glusterrest commands',
                                      metavar='')

    # Sub commands
    subparser.add_parser('install', help='Installation')
    subparser.add_parser('reinstall',
                         help='Reinstall - Note: Users database will be reset')
    parser_show = subparser.add_parser(
        'show',
        help='Show info about users/groups/config'
    )
    parser_port = subparser.add_parser('port', help='REST API PORT')
    parser_useradd = subparser.add_parser('useradd', help='Add REST user')
    parser_usermod = subparser.add_parser('usermod', help='Modify REST user')
    parser_userdel = subparser.add_parser('userdel', help='Delete REST user')
    parser_passwd = subparser.add_parser('passwd', help='Change Password')

    # Show arguments
    parser_show.add_argument('option',
                             type=str,
                             help='Show info about users/groups/config',
                             choices=['users', 'groups', 'config'])

    # Port arguments
    parser_port.add_argument('port',
                             type=int,
                             help='REST services PORT',
                             default=9000)

    # useradd arguments
    parser_useradd.add_argument('username',
                                type=str,
                                help='USERNAME')
    parser_useradd.add_argument('-g',
                                '--group',
                                type=str,
                                help='GROUP',
                                default='glusteruser')
    parser_useradd.add_argument('-p',
                                '--password',
                                type=str,
                                help='PASSWORD',
                                default='')

    # usermod arguments
    parser_usermod.add_argument('username',
                                type=str,
                                help='USERNAME')
    parser_usermod.add_argument('-g',
                                '--group',
                                type=str,
                                help='GROUP',
                                default='glusteruser',
                                choices=['glusterroot', 'glusteradmin',
                                         'glusteruser'])
    parser_usermod.add_argument('-p',
                                '--password',
                                type=str,
                                help='PASSWORD')

    # userdel arguments
    parser_userdel.add_argument('username',
                                type=str,
                                help='USERNAME')

    # passwd arguments
    parser_passwd.add_argument('username',
                               type=str,
                               help='USERNAME')
    parser_passwd.add_argument('-p',
                               '--password',
                               type=str,
                               help='PASSWORD',
                               default='')

    return parser.parse_args()
