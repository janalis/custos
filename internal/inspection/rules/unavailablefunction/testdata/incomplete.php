<?php
// Without a complete world (Composer packages installed and indexed), a
// polyfill may declare the function in code the index never saw.
str_contains('haystack', 'needle');
