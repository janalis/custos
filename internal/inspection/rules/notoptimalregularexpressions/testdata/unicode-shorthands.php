<?php
// Under /u the shorthands cover every script: the classes stay as written.
preg_match('/^-?[0-9]+(\.[0-9]+)?$/u', $number);
preg_match('/^[A-Za-z0-9_]+$/u', $name);
preg_match('/[^0-9]/u', $text);
// A negated shorthand means the same in every mode.
preg_match(<weak_warning descr="Write '[^\s]' as '\S' (same result).">'/[^\s]+/u'</weak_warning>, $text);
// Without /u they are the same.
preg_match(<weak_warning descr="Write '[0-9]' as '\d' (same result without /u).">'/^[0-9]+$/'</weak_warning>, $number);
