#!/usr/bin/env python


import os, sys, glob, time

def runCmd(cmd):
    print cmd
    os.system(cmd)

N_COLORS_LIST = [4, 6, 8, 10, 20]

inFns = glob.glob('input/*')
for ii, inFn in enumerate(inFns):
    print
    print '--- %s --- (%s/%s)' % (inFn, ii, len(inFns))
    for N_COLORS in N_COLORS_LIST:
        print
        print '%s colors' % N_COLORS
        tempFn = inFn.split('/')[-1].rsplit('.',1)[0] + '.reduced_color.png'
        sortFn = 'output/' + inFn.split('/')[-1].rsplit('.',1)[0] + '.reduced_color.congregated.png'
        finalFn = 'output/' + inFn.split('/')[-1].rsplit('.',1)[0] + '.reduced_color.congregated_%02i.png' % N_COLORS
        if os.path.exists(finalFn):
            continue
        try:
            runCmd('convert "%s" +dither -colors %s "%s"' % (inFn, N_COLORS, tempFn))
            runCmd('./sort-pixels "%s"' % tempFn)
        finally:
            runCmd('mv "%s" "%s"' % (sortFn, finalFn))
            runCmd('rm -f "%s"' % tempFn)
        time.sleep(0.5)


#
