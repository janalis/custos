<?php
namespace Http;

function parse_str($raw) { return []; }

function readQuery($raw)
{
    parse_str($raw);
    \<error descr="Pass a result array as second argument instead of creating variables.">PARSE_STR</error>($raw);
}
