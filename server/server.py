#!/usr/bin/env python

import time
import ujson
import asyncio
import datetime
import websockets

CONNECTED_CLIENTS = []

auto_correct = {'engineer': 'car mechanic',
                'bitcoin': 'magic bean',
                'java': 'jibberish'}

async def handle_message(encoded_message):
    # Grab message time ** May want to move this to client side **
    message_received_time = datetime.datetime.now().strftime("%H:%M")

    decoded_message = ujson.loads(encoded_message)

    id = decoded_message['id']
    name = decoded_message['name']
    message = decoded_message['message']

    msg = {'id': id,
           'name': name,
           'message': message,
           'time': message_received_time}
    msg = ujson.dumps(msg)
    for ws in CONNECTED_CLIENTS:
        await ws.send(msg)
    print(msg)
# end def handle_message


async def connection(websocket, path):
    global CONNECTED_CLIENTS

    # Grab user connection time
    start_time = datetime.datetime.now().strftime("%H:%M")

    # Add newly connected client to global connection list
    CONNECTED_CLIENTS.append(websocket)

    # Wait for client name to be passed through and extract username
    login = await websocket.recv()
    name = ujson.loads(login)['name']
    id = ujson.loads(login)['id']

    #for ws in CONNECTED_CLIENTS:
    #    await ws.send(str(start_server) + '  -  ' + str(name) + " Joined the Conversation")
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
            #for ws in CONNECTED_CLIENTS:
            #    await ws.send(end_time + '  -  ' + name + " Left the Conversation")
            print(end_time + '  -  ' + name + " Left the Conversation")
            break
# end def connection

start_server = websockets.serve(connection, '127.0.0.1', 9898)
asyncio.get_event_loop().run_until_complete(start_server)
asyncio.get_event_loop().run_forever()
