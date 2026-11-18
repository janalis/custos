<?php
if (Fiber::getCurrent() !== null) { Fiber::suspend(); }
