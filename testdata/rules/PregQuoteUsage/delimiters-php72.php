<?php
function find(string $ext, string $name)
{
    // '#' is escaped by preg_quote() from PHP 7.3 only.
    return preg_match('#^' . <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($ext) . '#', $name);
}
