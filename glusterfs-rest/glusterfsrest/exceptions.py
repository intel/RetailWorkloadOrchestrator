# -*- coding: utf-8 -*-
"""
    exceptions.py

    :copyright: (c) 2014 by Aravinda VK
    :license: MIT, see LICENSE for more details.
"""

import defusedxml.cElementTree as etree


class GlusterCliFailure(Exception):
    pass


class GlusterCliBadXml(Exception):
    pass


ParseError = etree.ParseError if hasattr(etree, 'ParseError') else SyntaxError
