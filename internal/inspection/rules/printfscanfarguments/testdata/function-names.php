<?php
namespace Shop\Billing {
    function sprintf($format, ...$values) { return $format; }

    function invoice($id, $total)
    {
        // user function declared in this namespace: not the builtin
        echo sprintf('%s-%s-%s', $id);
        // fully qualified builtin, written in mixed case
        echo <error descr="This call needs 3 argument(s) in total.">\SPrintf</error>('%s/%s', $id);
        echo \Printf(<error descr="Malformed format string.">'100 %, done %s'</error>, $total);
    }
}

namespace Shop\Mail {
    use function Shop\Billing\sprintf;

    function subject($who)
    {
        // imported user function
        return sprintf('%s %s %s', $who);
    }
}

namespace {
    function receipt($fh, $line, $amount)
    {
        <error descr="This call needs 4 argument(s) in total.">FPrintF</error>($fh, '%s: %d', $line, $amount, 'extra');
        <error descr="This call needs 4 argument(s) in total.">SScanf</error>($line, '%d %d', $first);
        [$a, $b] = SSCANF($line, '%d %d');
        echo Vendor\sprintf('%s %s', $line);
    }
}
