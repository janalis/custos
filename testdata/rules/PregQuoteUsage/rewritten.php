<?php
function glob_to_regex(string $glob, string $dir, array $parts): array
{
    return [
        // Braces are escaped by preg_quote() itself, also after a rewrite.
        '{^' . str_replace('\\*', '.*', preg_quote($glob)) . '$}i',
        '{/*' . strtr(preg_quote($dir), ['/' => '/+']) . '/?$}',
        '/^' . str_replace('\\*', '.*', <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($glob)) . '$/',
        '{' . str_replace(<error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($glob), '', $dir) . '}',
        '{' . str_replace('a', 'b', subject: <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($glob)) . '}',
        '{' . trim(<error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($glob)) . '}',
        '{' . str_replace('a', <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($dir)) . '}',
    ];
}
