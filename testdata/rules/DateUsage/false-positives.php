<?php
function stamp($pattern, Clock $clock)
{
    echo date('H:i', time(42));
    echo date('H:i', $clock->time());
    echo date('H:i', Clock::time());
    echo date('H:i', (time()));
    echo gmdate('H:i', time());
    echo date('H:i');
    echo date('H:i', ...$pattern);
}
