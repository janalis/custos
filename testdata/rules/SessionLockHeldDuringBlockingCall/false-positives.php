<?php

// The session is read and closed at once: no lock is held afterwards.
function readOnly()
{
    ini_set('session.save_handler', 'files');
    session_start(['read_and_close' => true]);
    sleep(5);
}

// Options that cannot be read statically may close the session.
function dynamicOptions(array $options)
{
    ini_set('session.save_handler', 'files');
    session_start($options);
    sleep(5);
}

// A later setting replaces the files handler.
function handlerReplaced()
{
    ini_set('session.save_handler', 'files');
    ini_set('session.save_handler', 'redis');
    session_start();
    sleep(5);
}

function handlerMaybeReplaced(bool $shared)
{
    ini_set('session.save_handler', 'files');
    if ($shared) {
        ini_set('session.save_handler', 'memcached');
    }
    session_start();
    sleep(5);
}

function handlerRestored()
{
    ini_set('session.save_handler', 'files');
    ini_restore('session.save_handler');
    session_start();
    sleep(5);
}

// A user handler decides its own locking.
function customHandler(SessionHandlerInterface $handler)
{
    ini_set('session.save_handler', 'files');
    session_set_save_handler($handler, true);
    session_start();
    sleep(5);
}

function moduleSelected()
{
    ini_set('session.save_handler', 'files');
    session_module_name('redis');
    session_start();
    sleep(5);
}

function closedFirst($ch)
{
    ini_set('session.save_handler', 'files');
    session_start();
    session_write_close();
    curl_exec($ch);
}
