<?php $show = true; if ($show) {} ?><b>always printed</b><?php
while (false) {} ?><i>also printed</i><?php
if ($show): ?><u>alternative syntax is fine</u><?php endif;
