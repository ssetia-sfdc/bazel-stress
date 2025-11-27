def _remote_ex_rule(ctx):
    # Collect all input files (deps outputs + input_files if specified)
    input_files = ctx.files.deps
    if ctx.files.input_files:
        input_files = input_files + ctx.files.input_files
    
    arguments = [str(ctx.attr.usecs)]
    
    # Add all input file paths as arguments
    if ctx.files.input_files:
        for input_file in ctx.files.input_files:
            arguments.append(input_file.path)
    else:
        # If no input files, pass empty string
        arguments.append("")
    
    arguments.append(ctx.outputs.txt.path)
    
    ctx.actions.run(
        inputs = input_files,
        outputs = [ctx.outputs.txt],
        executable = ctx.executable._ex,
        arguments = arguments,
    )
    
remote_ex_rule = rule(
    implementation = _remote_ex_rule,
    attrs = {
        "deps": attr.label_list(),
        "usecs": attr.int(mandatory=True),
        "input_files": attr.label_list(allow_files=True, doc="Input files to read random lines from"),
        "_ex": attr.label(default="//tools:ex", executable=True, cfg="exec"),
    },
    outputs = {
        "txt": "%{name}.txt",
    }
)
