<?php
function good($password, $stored) { return password_verify($password, $stored); }
