#include "mupdf/fitz.h"
#include "mupdf/pdf.h"

#include <string.h>
#include <stdlib.h>
#include <stdio.h>

int pdfclean(char *infile, char *outfile)
{
	pdf_write_options opts = pdf_default_write_options;
	int errors = 0;
	fz_context *ctx;
	char *password = "";
	char **retainlist;

	ctx = fz_new_context(NULL, NULL, FZ_STORE_UNLIMITED);
	if (!ctx)
	{
		fprintf(stderr, "cannot initialise context\n");
		exit(1);
	}

	fz_try(ctx)
	{
		pdf_clean_file(ctx, infile, outfile, password, &opts, retainlist, 0);
	}
	fz_catch(ctx)
	{
		errors++;
	}
	fz_drop_context(ctx);

	return errors != 0;
}
