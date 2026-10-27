<?php
/**
 * @param string|resource $content
 */
function dump($content): void
{
    if (\is_array($content)) {
        throw new \TypeError('array given');
    }
    if (!(\is_string($content) || \is_resource($content))) throw new \TypeError('bad');
    if (!\is_object($content)) throw new \TypeError('object expected');
}

/**
 * @param string $s
 */
function loose($s): void
{
    if (<warning descr="This check is always false for the declared parameter type; is the parameter being reused?">\is_array($s)</warning>) {
        echo 1;
    }
    if (<warning descr="This check is always false for the declared parameter type; is the parameter being reused?">\is_array($s)</warning>) {
        throw new \TypeError('x');
    } else {
        echo 2;
    }
    if (<warning descr="This check is always false for the declared parameter type; is the parameter being reused?">\is_array($s)</warning>) {
        echo 3;
        throw new \TypeError('y');
    }
}

function native(string $s): void
{
    if (<warning descr="This check is always false for the declared parameter type; is the parameter being reused?">\is_array($s)</warning>) {
        throw new \TypeError('z');
    }
}
