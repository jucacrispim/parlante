.. parlante documentation master file, created by
   sphinx-quickstart on Tue Aug 19 00:04:59 2025.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

Parlante: web comments powered by sqlite
========================================

Parlante is a simple comments system written in go that uses sqlite
as a database.


Installation
------------

To install use:

.. code-block:: sh

   $ go install github.com/jucacrispim/parlante/cmd/parlante
   $ go install github.com/jucacrispim/parlante/cmd/parlante-tui


Usage
-----

Parlante consists of two parts: A text user interface where you manage
clients, domains and comments and a server to add/list comments in a web page.

Clients must authorize domains in order to comments be listed/added in pages
in a given domain. Create a user and add at least one domain using the tui.

.. code-block:: sh

   $ parlante-tui -dbpath /path/to/my/sqlite.db


.. note::

   During the creation of the client a key is shown in the screen. Save it
   as the key can't be recovered.

Now start the server with:

.. code-block:: sh

   $ parlante -dbpath /path/to/my/sqlite.db


Comments
~~~~~~~~

The parlante api has endpoints that return html for easy of use or json for
more control.

The easiest way to use parlante is simply include the parlante.js that calls
the html endpoint and includes the comments list and the comment form in the
page.

.. code-block:: html

   <div id="#comments-container"></div>
   <script src="<PARLANTE_URL>/parlante.js" async onload="parlanteLoadComments('<PARLANTE_URL>', '<CLIENT_UUID>', 'comments-container')"></script>


This is going to render the comments and the comment form in the page.

For the create comment and list comment json endpoints, check `post comment <./swagger/#/paths/~1comments~1/post>`_
and the `get comments <./swagger/#/paths/~1comments~1%7Bclient_uuid%7D/get>`_
endpoints.


Counting comments
~~~~~~~~~~~~~~~~~

You can also count the comments in a list of web pages. Use the
`count comments <./swagger/#/paths/~1comments~1{client_uuid}~1count/post>`_ endpoint.


Contact form
~~~~~~~~~~~~

There is also a contact form available. To the html version use:

.. code-block:: html

   <div id="contact-container"></div>
   <script src="<PARLANTE_URL>/parlante.js" async onload="parlanteGetPingMeForm('<PARLANTE_URL>', '<CLIENT_UUID>', 'contact-container')"></script>


For the js endpoints check the `pingme <./swagger/#/paths/~1pingme~1{client_uuid}/post>`_.
