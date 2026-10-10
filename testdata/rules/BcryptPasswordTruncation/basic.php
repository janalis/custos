<?php
password_hash(<warning descr="Keep bcrypt input within its supported byte length.">str_repeat("p",80)</warning>,PASSWORD_BCRYPT);
