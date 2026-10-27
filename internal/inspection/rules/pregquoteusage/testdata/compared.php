<?php
function localized(array $req, string $locale): bool
{
    return ($req['_locale'] ?? null) === preg_quote($locale)
        || (preg_quote($locale)) == $req['x']
        || $req['y'] !== preg_quote($locale)
        || $req['z'] != preg_quote($locale)
        || $req['w'] < <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($locale);
}
