# -*- coding: utf-8 -*-
"""
    restapp.py

    :copyright: (c) 2014 by Aravinda VK
    :license: MIT, see LICENSE for more details.
"""

import types
from flask import Flask, request, jsonify
from functools import wraps
from glusterfsrest import users
from glusterfsrest.config import APP_DEBUG
from glusterfsrest.exceptions import GlusterCliBadXml, GlusterCliFailure

app = Flask(__name__)
app.debug = APP_DEBUG
users.connect()


def resp_success(data):
    message = {
        'ok': True,
        'data': data
    }
    return jsonify(message)


def resp_error(code, error, headers=None):
    message = {
        'ok': False,
        'error': error
    }
    resp = jsonify(message)
    resp.status_code = code
    if headers:
        for k, v in headers.items():
            resp.headers[k] = v
    return resp


def run_and_response(func, args):
    """
    Gets func to execute and arguments,
    if execution is success then sends response
    as {ok: true, data: <DATA>}
    else {ok: false, error: <ERROR>}
    """
    try:
        return resp_success(func(*args))
    except (GlusterCliFailure, GlusterCliBadXml) as e:
        return resp_error(200, str(e).strip())


def get_post_data(key, default_value=None):
    val = request.form[key] if key in request.form else default_value
    if isinstance(default_value, bool):
        val = True if val else False

    return val


@app.errorhandler(404)
def not_found(error):
    return resp_error(404, str(error))


@app.errorhandler(403)
def forbidden(error):
    return resp_error(403, str(error))


@app.errorhandler(405)
def notsupported(error):
    return resp_error(405, str(error))


def authenticate_error():
    """Sends a 401 response that enables basic auth"""
    return resp_error(
        401,
        'Forbidden',
        headers={'WWW-Authenticate': 'Basic realm="Login Required"'}
    )


def requires_auth(groups=[]):
    def requires_auth_decorator(f):
        @wraps(f)
        def decorated(*args, **kwargs):
            auth = request.authorization
            if not auth or not users.authenticate(auth.username,
                                                  auth.password,
                                                  groups):
                return authenticate_error()
            return f(*args, **kwargs)
        return decorated
    return requires_auth_decorator
