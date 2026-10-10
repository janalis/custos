<?php
function bad($password, $stored) { if (password_verify($password, $stored)) { return true; } return false; }
