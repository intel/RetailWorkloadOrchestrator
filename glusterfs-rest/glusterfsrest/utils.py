# -*- coding: utf-8 -*-
"""
    utils.py

    :copyright: (c) 2014 by Aravinda VK
    :license: MIT, see LICENSE for more details.
"""

import subprocess
import defusedxml.cElementTree as etree
from glusterfsrest.exceptions import GlusterCliFailure, GlusterCliBadXml
from glusterfsrest.exceptions import ParseError


def execute(cmd, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
            stderr=subprocess.PIPE, env=None, close_fds=True):
    p = subprocess.Popen(cmd,
                         stdin=stdin,
                         stdout=stdout,
                         stderr=stderr,
                         env=env,
                         close_fds=close_fds)

    (out, err) = p.communicate()
    return (p.returncode, out, err)


def checkstatuszero(cmd):
    rc, _, err = execute(cmd)
    if rc == 0:
        return True

    raise GlusterCliFailure(err)


def execute_and_output(cmd, func):
    rc, out, err = execute(cmd + ['--xml'])
    if rc == 0:
        return func(out)

    raise GlusterCliFailure(err)


def checkxmlcorrupt(xmldata):
    try:
        return etree.fromstring(xmldata)
    except (ParseError, AttributeError, ValueError) as e:
        raise GlusterCliBadXml(str(e))
