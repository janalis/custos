<?php

function lockKeptOpen()
{
    ini_set('session.save_handler', 'files');
    session_start(['read_and_close' => false, 'cookie_lifetime' => 60]);
    <warning descr="Release the session lock before blocking when safe.">sleep(5)</warning>;
    <warning descr="Release the session lock before blocking when safe.">usleep(microseconds: 10)</warning>;
}

function filesSelectedLast($ch)
{
    ini_set('session.save_handler', 'redis');
    ini_set('session.save_handler', 'files');
    ini_set('session.gc_maxlifetime', '60');
    session_start();
    <warning descr="Release the session lock before blocking when safe.">curl_exec($ch)</warning>;
}
