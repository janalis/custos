<?php
function positional($name, $total)
{
    echo <error descr="This call needs 3 argument(s) in total.">sprintf</error>('%2$s %1$s', $name);
    echo <error descr="This call needs 2 argument(s) in total.">sprintf</error>("%1\$s and %1\$s again", $name, $total);
    echo <error descr="This call needs 2 argument(s) in total.">sprintf</error>('Price $amount: %s');
    echo sprintf(<error descr="Malformed format string.">'%1$s at 50 %'</error>, $name);

    echo sprintf('%2$s %1$s', $name, $total);
    echo sprintf('%1$s, %1$s!', $name);
    echo sprintf("%1\$05d \$%2\$s", $total, $name);
    echo sprintf('Price $amount: %s', $total);
    echo sprintf("Hi $name %1\$s", $total, 'x');
}
