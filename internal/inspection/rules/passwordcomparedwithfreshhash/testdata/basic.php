<?php
function bad($password, $stored) { if (<warning descr="Verify the password against the stored hash.">password_hash($password, PASSWORD_DEFAULT) === $stored</warning>) { return true; } return false; }
function literal($password) { return <warning descr="Verify the password against the stored hash.">password_hash($password, PASSWORD_DEFAULT) === '$2y$10$abcdefghijklmnopqrstuv'</warning>; }
