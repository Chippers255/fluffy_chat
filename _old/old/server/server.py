#!/usr/bin/env python

import time
import ujson
import asyncio
import datetime
import websockets

CONNECTED_CLIENTS = []
CLIENT_NAMES = {}

auto_correct = {'engineer': 'car mechanic',
                'bitcoin': 'magic bean',
                'java': 'jibberish'}

async def handle_message(encoded_message):
    message_received_time = datetime.datetime.now().strftime("%H:%M")

    decoded_message = ujson.loads(encoded_message)

    id = decoded_message['id']
    name = decoded_message['name']
    message = decoded_message[ 'message']

    msg = {'id': id,
           'name': name,
           'message': message,
           'time': message_received_time}

    msg = ujson.dumps(msg)

    for ws in CONNECTED_CLIENTS:
        await ws.send(msg)

    print(msg)
# end def handle_message


async def update_users():
    msg = {'type': 'update_users',
           'id': 'server',
           'users': []}

    for name in CLIENT_NAMES.values():
        msg['users'].append(name)

    msg = ujson.dumps(msg)

    for ws in CONNECTED_CLIENTS:
        await ws.send(msg)
# end def update_users

async def connection(websocket, path):
    global CONNECTED_CLIENTS, CLIENT_NAMES

    # Grab user connection time
    start_time = datetime.datetime.now().strftime("%H:%M")

    # Add newly connected client to global connection list
    CONNECTED_CLIENTS.append(websocket)

    # Wait for client name to be passed through and extract user connection info
    login = await websocket.recv()
    con_info = ujson.loads(login)
    type = con_info['type']
    id = con_info['id']
    name = con_info['name']

    # Add to list of usernames
    CLIENT_NAMES[id] = name

    await update_users()

    print(start_time + '  -  ' + name + " Joined the Conversation")  # Log to server

    # Loop waiting on client messages until disconnection
    while True:
        try:
            message = await websocket.recv()
            await handle_message(message)
        except:
            # Grab user disconnection time
            end_time = datetime.datetime.now().strftime("%H:%M")

            CONNECTED_CLIENTS.remove(websocket)
            del CLIENT_NAMES[id]

            await update_users()

            print(end_time + '  -  ' + name + " Left the Conversation")
            break
# end def connection

start_server = websockets.serve(connection, '127.0.0.1', 9898)
asyncio.get_event_loop().run_until_complete(start_server)
asyncio.get_event_loop().run_forever()
