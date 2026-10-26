<?php
function trimChars(string $s, string $chars): string
{
    $chars = preg_quote($chars);
    return preg_replace("{^[$chars]++|[$chars]++$}uD", '', $s);
}

function concatenated(string $s, string $word): string
{
    $word = (preg_quote($word));
    $fn = fn () => '#' . $word;
    $s = preg_replace('{' . $word . '}', '', $s);
    return preg_replace("( $word)", '', $s);
}

function laterSlash(string $s, string $word): string
{
    $word = <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($word);
    return preg_replace('/' . "{$word}" . '/', '', $s);
}

function rewritten(string $s, string $word): string
{
    $word = <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($word);
    $word = strtr($word, ['\*' => '.*']);
    return preg_replace("{{$word}}", '', $s);
}

function escaped(string $s, string $word): string
{
    $word = <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($word);
    return preg_replace("\x7b$word}", '', $s) . <<<EOT
    {$word}
    EOT;
}

function captured(string $s, string $word): callable
{
    $word = <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($word);
    return function () use ($word, $s) { return preg_replace("{{$word}}", '', $s); };
}

function unused(string $word): void
{
    $word = <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($word);
    $other = <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($word) . '|';
    $other .= <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($word);
    $this->word = <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($word);
    $$other = <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($word);
    echo $other = <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($word);
}

$global = <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($argv[1]);
echo preg_match("{{$global}}", 'x');
