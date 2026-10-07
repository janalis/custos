<?php
function strip(string $ext, string $name, string $sep, $obj): array
{
    return [
        // preg_quote() escapes these delimiters itself.
        preg_replace('|' . preg_quote($ext) . '$|', '', $name),
        preg_replace(('![') . preg_quote($sep) . ']+!u', $sep, $name),
        preg_match(" {^" . (preg_quote($ext) . '}'), $name),
        preg_match('#^' . preg_quote($ext) . '#', $name),
        preg_match(sprintf('~^%s~', 'x') . sprintf('(%s)', preg_quote($ext)), $name),
        preg_match(sprintf('(^%s)', preg_quote($ext)), $name),
        // '/' and '~' are not escaped, or the delimiter is not visible.
        preg_match('/^' . <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($ext) . '/', $name),
        preg_match('~' . <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($ext) . '~', $name),
        preg_match($sep . <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($ext), $name),
        preg_match('' . <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($ext), $name),
        preg_match(sprintf('/^%s/', <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($ext)), $name),
        preg_match(sprintf(<error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($ext)), $name),
        preg_match($obj->format('|%s|', <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($ext)), $name),
        preg_match(sprintf(...$sep, <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($ext)), $name),
        <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($ext),
    ];
}
