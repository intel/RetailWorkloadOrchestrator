# -*- coding: utf-8 -*-
"""
    cli.volume.py

    :copyright: (c) 2014 by Aravinda VK
    :license: MIT, see LICENSE for more details.
"""

from glusterfsrest import utils
from glusterfsrest.exceptions import GlusterCliBadXml, ParseError
from glusterfsrest.exceptions import GlusterCliFailure


VOLUME_CMD = ['gluster', '--mode=script', 'volume']


def _parse_a_vol(volume_el):
    value = {
        'name': volume_el.find('name').text,
        'uuid': volume_el.find('id').text,
        'type': volume_el.find('typeStr').text.upper().replace('-', '_'),
        'status': volume_el.find('statusStr').text.upper(),
        'num_bricks': int(volume_el.find('brickCount').text),
        'distribute': int(volume_el.find('distCount').text),
        'stripe': int(volume_el.find('stripeCount').text),
        'replica': int(volume_el.find('replicaCount').text),
        'transport': volume_el.find('transport').text,
        'bricks': [],
        'options': []
    }
    if value['transport'] == '0':
        value['transport'] = 'TCP'
    elif value['transport'] == '1':
        value['transport'] = 'RDMA'
    else:
        value['transport'] = 'TCP,RDMA'

    for b in volume_el.findall('bricks/brick'):
        try:
            value['bricks'].append({"name": b.find("name").text,
                                    "hostUuid": b.find("hostUuid").text})
        except AttributeError:
            value['bricks'].append(b.text)

    for o in volume_el.findall('options/option'):
        value['options'].append({"name": o.find('name').text,
                                 "value": o.find('value').text})

    return value


def _parseinfo(volinfo):
    tree = utils.checkxmlcorrupt(volinfo)
    volumes = []
    for el in tree.findall('volInfo/volumes/volume'):
        try:
            volumes.append(_parse_a_vol(el))
        except (ParseError, AttributeError, ValueError) as e:
            raise GlusterCliBadXml(str(e))

    return volumes


def info(name=None):
    cmd = VOLUME_CMD + ["info"] + ([name] if name else [])
    data = utils.execute_and_output(cmd, _parseinfo)
    if name and not data:
        raise GlusterCliFailure("Volume %s does not exist" % name)

    return data


def start(name, force=False):
    cmd = VOLUME_CMD + ["start", name]
    if force:
        cmd += ["force"]

    return utils.checkstatuszero(cmd)


def stop(name, force=False):
    cmd = VOLUME_CMD + ["stop", name]
    if force:
        cmd += ["force"]

    return utils.checkstatuszero(cmd)


def create(name, bricks, replica=0, stripe=0, transport='tcp', force=False,
           start_volume=False):
    cmd = VOLUME_CMD + ["create", name]
    if stripe:
        cmd += ["stripe", str(stripe)]

    if replica:
        cmd += ["replica", str(replica)]

    cmd += ["transport", transport]

    cmd += bricks

    if force:
        cmd += ["force"]

    # If volume needs to be started, then run create command without
    # decorator else return create command and statuszerotrue
    # decorator will take care of running cmd
    if start_volume:
        utils.checkstatuszero(cmd)
        return start(name, force=True)
    else:
        return utils.checkstatuszero(cmd)


def delete(name, stop_volume=False):
    if stop_volume:
        stop(name, force=True)

    cmd = VOLUME_CMD + ["delete", name]
    return utils.checkstatuszero(cmd)


def restart(name):
    stop(name, force=True)
    start(name)
    return True


def addbrick(name, brickpath, replica=0, stripe=0, force=False):
    cmd = VOLUME_CMD + ["add-brick", name]
    if stripe:
        cmd += ["stripe", stripe]

    if replica:
        cmd += ["replica", replica]

    cmd += [brickpath]
    if force:
        cmd += ["force"]

    return utils.checkstatuszero(cmd)


def removebrickForce(name, brickpath, replica=0):
    cmd = VOLUME_CMD + ["remove-brick", name]

    if replica:
        cmd += ["replica", replica]

    cmd += [brickpath, "force"]

    return utils.checkstatuszero(cmd)


def removebrickStart(name, brickpath, replica=0):
    cmd = VOLUME_CMD + ["remove-brick", name]

    if replica:
        cmd += ["replica", replica]

    cmd += [brickpath, "start"]

    return utils.checkstatuszero(cmd)


def removebrickStop(name, brickpath, replica=0):
    cmd = VOLUME_CMD + ["remove-brick", name]

    if replica:
        cmd += ["replica", replica]

    cmd += [brickpath, "stop"]

    return utils.checkstatuszero(cmd)


def removebrickCommit(name, brickpath, replica=0):
    cmd = VOLUME_CMD + ["remove-brick", name]

    if replica:
        cmd += ["replica", replica]

    cmd += [brickpath, "commit"]

    return utils.checkstatuszero(cmd)
