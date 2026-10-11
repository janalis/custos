<?php
function fill($target)
{
}

final class CallTime
{
    // Call-time pass-by-reference (PHP 5.3) binds whatever the callee declares.
    public function filled()
    {
        $r = '';
        fill(&$r);
        return $r;
    }

    public function <weak_warning descr="Declare ': int' as the return type.">orThrow</weak_warning>($flag)
    {
        if ($flag) {
            return 1;
        }
        throw new \RuntimeException('none');
    }
}
