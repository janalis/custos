<?php
function bad($password, $stored) { if (password_hash($password, PASSWORD_DEFAULT) === $stored) { return true; } return false; }
