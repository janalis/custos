<?php

function matchedFirst($path)
{
    $code = file_get_contents($path);
    if (!<warning descr="Handle file reading failure before using its result.">preg_match('/class\s+(\w+)/', $code, $m)</warning>) {
        return '';
    }

    return strlen($code) . $m[1];
}

function matchedInBranch($path)
{
    $code = file_get_contents($path);
    if (<warning descr="Handle file reading failure before using its result.">preg_match_all('#\d+#u', $code)</warning>) {
        return trim($code);
    }

    return '';
}

function emptyMatch($path)
{
    $code = file_get_contents($path);
    if (<warning descr="Handle file reading failure before using its result.">preg_match('/a*/', $code)</warning>) {
        return <warning descr="Handle file reading failure before using its result.">strlen($code)</warning>;
    }

    return 0;
}

function unreadPattern($path)
{
    $code = file_get_contents($path);
    if (<warning descr="Handle file reading failure before using its result.">preg_match('/(?=a)b/x', $code)</warning>) {
        return <warning descr="Handle file reading failure before using its result.">strlen($code)</warning>;
    }

    return 0;
}
