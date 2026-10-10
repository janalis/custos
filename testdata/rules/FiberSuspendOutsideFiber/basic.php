<?php
if (Fiber::getCurrent() === null) { <error descr="Suspend only inside a running fiber.">Fiber::suspend()</error>; }
