<?php
// A `...` placeholder after arguments is a parse error: not rewritten.
call_user_func($hook, $to, ...);
call_user_func_array($hook, ...);
