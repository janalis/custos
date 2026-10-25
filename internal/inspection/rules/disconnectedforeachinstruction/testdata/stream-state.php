<?php
function replay($data, $lock, array $items, callable $onWarning)
{
    foreach ($items as $item) {
        ftruncate($data, 0);
        rewind($data);
        fseek($data, 0, SEEK_END);
        fread($data, 1);
        flock($lock, LOCK_UN);
        set_error_handler($onWarning);
        handle($item);
        restore_error_handler();
        <weak_warning descr="Statement does not depend on the loop; move it out.">fstat($data);</weak_warning>
    }
}
