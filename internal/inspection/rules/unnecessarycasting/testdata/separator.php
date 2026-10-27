<?php
function separators(int $b, string $s, array $items)
{
    $x = 10 -(int)-$b;
    $y = 1 +<weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning>+$b;
    $z = $s.<weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning>2;
    $w = $s.<weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning>.5;
    print<weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning>5;
    echo<weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning>7, "\n";
    switch ($b) {
        case<weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning>3:
            break;
    }
    $v = [<weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning>$b,<weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning>$b, !<weak_warning descr="Operand already has the target type; remove the cast.">(bool)</weak_warning>true];
    $u =(int)-$b;
    $t = $b ?<weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning>$b :<weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $b;
    return<weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning>count($items);
}
