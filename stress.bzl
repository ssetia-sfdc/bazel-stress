def _remote_ex_rule(ctx):
    # Collect all input files (deps outputs + input_file if specified)
    input_files = ctx.files.deps
    if ctx.file.input_file:
        input_files = input_files + [ctx.file.input_file]
    
    arguments = [str(ctx.attr.usecs)]
    if ctx.file.input_file:
        arguments.append(ctx.file.input_file.path)
    else:
        # If no input file, create a default one on the fly
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
        "input_file": attr.label(allow_single_file=True, doc="Input file to read random lines from"),
        "_ex": attr.label(default="//tools:ex", executable=True, cfg="exec"),
    },
    outputs = {
        "txt": "%{name}.txt",
    }
)
