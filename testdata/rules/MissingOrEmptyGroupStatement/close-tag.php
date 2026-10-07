<?php $show = true; <weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> ($show) ?><b>always printed</b><?php
<weak_warning descr="Use a braced block for the body of this construct.">while</weak_warning> (false) ?><i>also printed</i><?php
if ($show): ?><u>alternative syntax is fine</u><?php endif;
