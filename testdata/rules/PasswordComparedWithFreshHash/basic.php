<?php
function bad($password, $stored) { if (<warning descr="Verify the password against the stored hash.">password_hash($password, PASSWORD_DEFAULT) === $stored</warning>) { return true; } return false; }
