<?php
try { throw new RuntimeException("a"); } finally { <warning descr="Preserve the original exception during cleanup.">throw new LogicException("b")</warning>; }
