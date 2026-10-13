<?php
function readQuery($raw)
{
    <error descr="Pass a result array as second argument instead of creating variables.">parse_str</error>($raw);
    \<error descr="Pass a result array as second argument instead of creating variables.">mb_parse_str</error> ($raw);
    <error descr="Pass a result array as second argument instead of creating variables.">Parse_Str</error>($raw);
}
