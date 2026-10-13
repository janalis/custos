<?php
enum Phase { case Fresh; } <error descr="Use an enum case instead of instantiation.">new Phase()</error>;
