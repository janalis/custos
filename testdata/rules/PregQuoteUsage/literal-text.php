<?php

function legacyPatterns(string $name): array
{
    return [
        '/[a-z_' . preg_quote('\\') . ']*(' . preg_quote('::') . ')/',
        '/' . preg_quote("Mage Core\t-1.0") . '/',
        '/' . <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>('a/b') . '/',
        '/' . <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>("x$name") . '/',
        '/' . <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>('a~b') . '/',
    ];
}
