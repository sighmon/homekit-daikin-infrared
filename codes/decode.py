import sys

filename = sys.argv[1]


def parse(input_string):
    """
    Attempt to parse LIRC output to a binary string.
    """
    def pulse_width_to_integer(pulse_width):
        """
        Convert the LIRC pulse width to an integer so we can create
        a lookup dictionary.
        """
        if pulse_width < 1000:
            return 0
        elif pulse_width < 1700:
            return 1
        elif pulse_width < 5000:
            return 2
        else:
            return 3

    # A = start of line
    # B = end of line
    # E = end of command
    tuple_to_binary = {(0,0):"0",(0,1):"1",(2,2):"A",(0,3):"B",(0,2):"E"}
    list_of_tuples = list(map(lambda x: int(x[6:]), input_string.split("\n")))
    binary_values = []

    for i in range(0, len(list_of_tuples), 2):
        binary_values.append(
            tuple_to_binary.get(
                (
                    pulse_width_to_integer(list_of_tuples[i]),
                    pulse_width_to_integer(list_of_tuples[i+1]),
                ),
                str((list_of_tuples[i], list_of_tuples[i+1])),
            )
        )
    binary_string = "".join(binary_values)

    return binary_string

with open(filename, "r") as file:
    # Remove the last two newline characters
    data = file.read()[:-2]

    # LIRC outputs a timeout rather than a space for the Daikin remote
    # so let's replace it
    data = data.replace("timeout", "space")

    print(parse(data))
