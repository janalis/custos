<?php
function constant($password) { return <warning descr="Verify the password against the stored hash.">password_hash($password, PASSWORD_DEFAULT) == STORED_HASH</warning>; }
function later($password, $stored) { $hash = password_hash($password, PASSWORD_DEFAULT); return <warning descr="Verify the password against the stored hash.">$stored !== $hash</warning>; }
