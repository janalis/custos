<?php
ini_set('session.save_handler', 'files'); session_start(); <warning descr="Release the session lock before blocking when safe.">sleep(8)</warning>; session_write_close();
