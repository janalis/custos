<?php
// Before PHP 8.0 null is coerced to a string and renames the session.
session_start(); <warning descr="Set the session name before starting the session.">session_name(null)</warning>;
