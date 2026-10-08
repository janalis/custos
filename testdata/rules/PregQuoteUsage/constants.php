<?php
// Constants holding delimiter-free literals.
final class Packer
{
    const PREFIX = 'Moodle archive file index  Count: ';
    const SLASHED = 'a/b';
}

const PLAIN = 'abc';

function countOf(string $s)
{
    return [
        preg_match('~^' . preg_quote(Packer::PREFIX) . '([0-9]+)~', $s),
        preg_match('~' . preg_quote(PLAIN) . '~', $s),
        preg_match('/' . <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>(Packer::SLASHED) . '/', $s),
        preg_match('/' . <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>(Packer::MISSING) . '/', $s),
        preg_match('/' . <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>(PHP_EOL) . '/', $s),
    ];
}
