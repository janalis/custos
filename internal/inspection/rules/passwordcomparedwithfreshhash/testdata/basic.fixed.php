<?php
function bad($password, $stored) { if (password_verify($password, $stored)) { return true; } return false; }
function literal($password) { return password_verify($password, '$2y$10$abcdefghijklmnopqrstuv'); }
